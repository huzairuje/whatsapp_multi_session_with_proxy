package handler

import (
	"html/template"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	log "github.com/sirupsen/logrus"
)

// For simplicity, jwtKey is defined here. In a real app, this should come from config.
var jwtKey = []byte("my_super_secret_key_please_change_in_production")
const tokenExpirationDuration = 1 * time.Hour // Token valid for 1 hour
const authTokenCookieName = "auth_token"

// Claims struct for JWT token
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// ServeLoginPage renders the login page or handles login attempts.
// For GET requests, it serves the dashboard/index.html file.
// For POST requests, it handles login logic.
func (h Handler) ServeLoginPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodGet:
			// Note: Using template.ParseFiles and Execute as per "parsing on each request is fine".
			// To use c.HTML as typically intended by Gin, templates would be loaded at router setup
			// (e.g., router.LoadHTMLGlob("dashboard/*")) and then c.HTML(http.StatusOK, "index.html", nil) would be called.
			tmpl, err := template.ParseFiles("dashboard/index.html")
			if err != nil {
				log.Errorf("Error parsing login page template: %v", err)
				c.String(http.StatusInternalServerError, "Error parsing login page template: %v", err)
				return
			}
			err = tmpl.Execute(c.Writer, nil)
			if err != nil {
				log.Errorf("Error executing login page template: %v", err)
				c.String(http.StatusInternalServerError, "Error executing login page template: %v", err)
			}
		case http.MethodPost:
			username := c.PostForm("username")
			password := c.PostForm("password")

			// Placeholder login logic
			if username == "admin" && password == "password" {
				log.Infof("Login successful for user: %s", username)

				expirationTime := time.Now().Add(tokenExpirationDuration)
				claims := &Claims{
					Username: username,
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(expirationTime),
						IssuedAt:  jwt.NewNumericDate(time.Now()),
						Issuer:    "whatsapp_multi_session_dashboard",
					},
				}

				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, err := token.SignedString(jwtKey)
				if err != nil {
					log.Errorf("Error signing token: %v", err)
					c.String(http.StatusInternalServerError, "Internal server error (token generation)")
					return
				}

				http.SetCookie(c.Writer, &http.Cookie{
					Name:     authTokenCookieName,
					Value:    tokenString,
					Expires:  expirationTime,
					Path:     "/dashboard", // Restrict cookie to dashboard paths
					HttpOnly: true,
					Secure:   c.Request.TLS != nil, // Send only over HTTPS if available
					SameSite: http.SameSiteLaxMode,
				})

				log.Infof("Token generated and cookie set for user: %s", username)
				c.Header("HX-Redirect", "/dashboard/home")
				c.Redirect(http.StatusFound, "/dashboard/home") // Standard redirect as fallback
			} else {
				log.Warnf("Login failed for user: %s", username)
				// Re-render login page with an error message or send a specific HTML fragment.
				// For simplicity, returning 401. HTMX can swap this into an error div.
				// To make this more user-friendly with HTMX, you'd return a snippet like:
				// <div id="error-message" class="error">Invalid username or password</div>
				// And the form would have hx-target="#error-message" hx-swap="innerHTML" for errors.
				c.String(http.StatusUnauthorized, "Invalid username or password. For HTMX, this response might need to be an HTML snippet.")
			}
		default:
			c.String(http.StatusMethodNotAllowed, "Method not allowed")
		}
	}
}

// ServeHomePage renders the dashboard home page.
// It parses dashboard/base.html and dashboard/home.html,
// then renders home.html within base.html.
func (h Handler) ServeHomePage() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Placeholder authentication: In a real app, check for a valid session token from a cookie.
		// For now, we assume if they reach here, they are "authenticated".

		// Note: Using template.ParseFiles and ExecuteTemplate as per "parsing on each request is fine".
		// To use c.HTML as typically intended by Gin, templates would be loaded at router setup
		// (e.g., router.LoadHTMLGlob("dashboard/*")) and then c.HTML(http.StatusOK, "base.html", gin.H{"Title": "Dashboard Home"})
		// with home.html's content being included via template inheritance/blocks.
		tmpl, err := template.ParseFiles("dashboard/base.html", "dashboard/home.html")
		if err != nil {
			log.Errorf("Error parsing dashboard templates: %v", err)
			c.String(http.StatusInternalServerError, "Error parsing dashboard templates: %v", err)
			return
		}

		// The `home.html` defines a "content" template.
		// We execute the "base.html" template, which should internally call {{template "content" .}}
		// The data (nil here) is passed to "base.html", which then passes it to "content".
		err = tmpl.ExecuteTemplate(c.Writer, "base.html", nil)
		if err != nil {
			log.Errorf("Error executing dashboard home template: %v", err)
			c.String(http.StatusInternalServerError, "Error executing dashboard home template: %v", err)
		}
	}
}

// ServeDashboardContent dynamically serves content for the dashboard pages like messages, devices, settings.
// This is an example for HTMX loading content into the #content div of base.html.
func (h Handler) ServeDashboardContent() gin.HandlerFunc {
    return func(c *gin.Context) {
        page := c.Param("page") // e.g., "messages.html", "devices.html"
        
        // Basic security: ensure only .html files from dashboard are served
        // In a real app, you'd have more robust checks or specific handlers for each page.
        if !template.HTMLEscapeString(page) == page || page == "" {
             log.Warnf("Attempted to access invalid page: %s", page)
             c.String(http.StatusBadRequest, "Invalid page requested.")
             return
        }

        templateFile := "dashboard/" + page
        
        // Check if file exists (optional, template.ParseFiles will error anyway)
        // _, err := os.Stat(templateFile)
        // if os.IsNotExist(err) {
        //    log.Errorf("Dashboard content file not found: %s", templateFile)
        //    c.String(http.StatusNotFound, "Content not found.")
        //    return
        // }

        tmpl, err := template.ParseFiles(templateFile)
        if err != nil {
            log.Errorf("Error parsing dashboard content template '%s': %v", templateFile, err)
            c.String(http.StatusInternalServerError, "Error loading content: %v", err)
            return
        }

        // Here, we execute the *specific* template that was requested (e.g., "messages.html")
        // Assumes these templates are self-contained or define their own necessary blocks.
        // If they are meant to be *within* base.html's content, this approach is slightly different
        // from ServeHomePage. ServeHomePage renders base.html which *includes* home.html.
        // For HTMX partials, you often just render the partial itself.
        err = tmpl.Execute(c.Writer, nil) 
        if err != nil {
            log.Errorf("Error executing dashboard content template '%s': %v", templateFile, err)
            c.String(http.StatusInternalServerError, "Error rendering content: %v", err)
        }
    }
}

// GetSentMessages returns a placeholder HTML for sent messages.
func (h Handler) GetSentMessages() gin.HandlerFunc {
	return func(c *gin.Context) {
		htmlContent := "<p>Sent messages will be displayed here. (Backend integration pending)</p>"
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlContent))
	}
}

// HandleDashboardSendMessage handles sending a single text message from the dashboard.
func (h Handler) HandleDashboardSendMessage() gin.HandlerFunc {
	return func(c *gin.Context) {
		senderJIDStr := c.PostForm("sender_jid")
		recipientJIDStr := c.PostForm("recipient_jid")
		message := c.PostForm("message")

		// Basic validation
		if senderJIDStr == "" || recipientJIDStr == "" || message == "" {
			responseHTML := "<p class='error'>Sender, Recipient JID, and Message cannot be empty.</p>"
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(responseHTML)) // OK for HTMX, error is in content
			return
		}

		// Parse sender_jid string (e.g., "user@s.whatsapp.net") into user and server parts
		parts := strings.Split(senderJIDStr, "@")
		if len(parts) != 2 {
			responseHTML := "<p class='error'>Invalid Sender JID format. Expected user@server.</p>"
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(responseHTML))
			return
		}
		senderUser := parts[0]
		senderServer := parts[1]
		
		// Validate that senderUser and senderServer are not empty after split
		if senderUser == "" || senderServer == "" {
			responseHTML := "<p class='error'>Invalid Sender JID format. User or server part is missing.</p>"
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(responseHTML))
			return
		}

		senderJID := types.NewJID(senderUser, senderServer)

		// Call the existing command handler
		// HandleSendTextMessage expects recipientJIDStr as string, which is convenient.
		messageID, err := h.CommandHandler.HandleSendTextMessage(senderJID, message, recipientJIDStr)
		if err != nil {
			log.Errorf("HandleDashboardSendMessage: Error sending message from %s to %s: %v", senderJIDStr, recipientJIDStr, err)
			responseHTML := fmt.Sprintf("<p class='error'>Failed to send message: %v</p>", err)
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(responseHTML))
			return
		}

		log.Infof("HandleDashboardSendMessage: Message sent successfully from %s to %s. ID: %s", senderJIDStr, recipientJIDStr, messageID)
		responseHTML := fmt.Sprintf("<p class='success'>Message sent successfully! ID: %s</p>", messageID)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(responseHTML))
	}
}

// GetActiveSenders returns an HTML string of <option> tags for active (logged-in) devices.
func (h Handler) GetActiveSenders() gin.HandlerFunc {
	return func(c *gin.Context) {
		devices := h.CommandHandler.HandleGetAllDevices(c.Request.Context())
		var options strings.Builder // Use strings.Builder for efficient string concatenation

		activeDeviceFound := false
		if devices != nil {
			for _, device := range devices {
				if device.IsLoggedIn {
					activeDeviceFound = true
					jid := device.User // User field usually contains the number part of JID
					if device.Server != "" { // Server might be empty for some custom setups, though unlikely for s.whatsapp.net
						jid += "@" + device.Server
					}
					displayName := device.PushName
					if displayName == "" {
						displayName = jid // Fallback to JID if PushName is empty
					}
					options.WriteString(fmt.Sprintf("<option value='%s'>%s</option>", jid, displayName))
				}
			}
		}

		if !activeDeviceFound {
			options.WriteString("<option value='' disabled>No active devices found</option>")
		}

		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(options.String()))
	}
}

// AuthMiddleware validates JWT token from cookie.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie(authTokenCookieName)
		if err != nil {
			log.Warnf("AuthMiddleware: Cookie '%s' not found: %v", authTokenCookieName, err)
			// If request is for an HTMX partial, returning a full page redirect might not be what HTMX expects.
			// It might be better to return a 401 and have HTMX handle it on the client-side,
			// possibly by redirecting or showing a message.
			// For now, a standard redirect to login for non-HTMX or full page loads.
			// For HTMX, could send HX-Redirect header.
			if c.GetHeader("HX-Request") == "true" {
				c.Header("HX-Redirect", "/dashboard/login")
				c.AbortWithStatus(http.StatusUnauthorized)
			} else {
				c.Redirect(http.StatusFound, "/dashboard/login")
				c.Abort()
			}
			return
		}

		tokenString := cookie
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid // Or a more specific error
			}
			return jwtKey, nil
		})

		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				log.Warnf("AuthMiddleware: Invalid token signature: %v", err)
				c.Redirect(http.StatusFound, "/dashboard/login") // Or c.AbortWithStatusJSON for API like behavior
				c.Abort()
				return
			}
			log.Warnf("AuthMiddleware: Error parsing token: %v", err)
			// Handle other errors like expired token
			if verr, ok := err.(*jwt.ValidationError); ok {
				if verr.Errors&jwt.ValidationErrorExpired != 0 {
					log.Info("AuthMiddleware: Token has expired.")
					// Clear the expired cookie
					http.SetCookie(c.Writer, &http.Cookie{
						Name:     authTokenCookieName,
						Value:    "",
						Path:     "/dashboard",
						Expires:  time.Unix(0, 0), // Expire immediately
						HttpOnly: true,
					})
				}
			}
			c.Redirect(http.StatusFound, "/dashboard/login")
			c.Abort()
			return
		}

		if !token.Valid {
			log.Warn("AuthMiddleware: Token is invalid.")
			c.Redirect(http.StatusFound, "/dashboard/login")
			c.Abort()
			return
		}

		log.Infof("AuthMiddleware: Token validated successfully for user %s", claims.Username)
		c.Set("username", claims.Username) // Pass username to subsequent handlers
		c.Next()
	}
}

// GetDeviceCount returns the count of logged-in devices.
func (h Handler) GetDeviceCount() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Assuming h.CommandHandler is properly initialized.
		// If CommandHandler could be nil, a check would be needed:
		// if h.CommandHandler == nil { // Or appropriate interface nil check
		//     log.Error("GetDeviceCount: CommandHandler is not initialized.")
		//     c.String(http.StatusInternalServerError, "Error: CommandHandler not available.")
		//     return
		// }

		devices := h.CommandHandler.HandleGetAllDevices(c.Request.Context())
		loggedInCount := 0
		if devices != nil { // Check if the slice itself is nil
			for _, device := range devices {
				// Assuming primitive.Device struct has an IsLoggedIn boolean field
				if device.IsLoggedIn {
					loggedInCount++
				}
			}
		} else {
			log.Warn("GetDeviceCount: HandleGetAllDevices returned nil slice.")
			// loggedInCount remains 0, which is appropriate.
		}

		htmlContent := fmt.Sprintf("<p>Logged-in Devices: %d</p>", loggedInCount)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlContent))
	}
}

// GetMessageCountGraphic returns a placeholder HTML for a message count graphic.
func (h Handler) GetMessageCountGraphic() gin.HandlerFunc {
	return func(c *gin.Context) {
		htmlContent := "<p>Message count graphic will be displayed here. (Backend integration pending)</p>"
		// Example of how you might include a simple SVG or an img tag pointing to a generated graphic
		// htmlContent := `
		//  <div>
		//      <h4>Messages Sent Over Time</h4>
		//      <svg width="300" height="100" xmlns="http://www.w3.org/2000/svg">
		//          <rect x="10" y="10" width="50" height="80" fill="blue" />
		//          <rect x="70" y="30" width="50" height="60" fill="green" />
		//          <rect x="130" y="50" width="50" height="40" fill="red" />
		//          <text x="10" y="95" font-family="Verdana" font-size="10">Mockup Graphic</text>
		//      </svg>
		//      <p>(This is a placeholder graphic)</p>
		//  </div>
		// `
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlContent))
	}
}
