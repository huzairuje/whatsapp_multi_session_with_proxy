package sessions

import (
	"crypto/rand"
	"encoding/hex"
	"log" // For logging warnings
	"net/http"
	"os" // For environment variable

	"github.com/gorilla/sessions"
)

// Store will hold the session store.
var Store *sessions.CookieStore

// SessionKey is the name of the cookie used to store the session.
const SessionKey = "dashboard-session"
const UserIDKey = "user_id" // Key for storing user ID in session

func InitSessionStore() {
	// Attempt to get session key from environment variable for production
	// For development, use a hardcoded key but log a warning
	authKey := os.Getenv("SESSION_AUTHENTICATION_KEY")
	if authKey == "" {
		// Generate a random 32-byte key for development if not set
		// In a real app, this should be a persistent key for production
		key := make([]byte, 32)
		_, err := rand.Read(key)
		if err != nil {
			// Fallback to a hardcoded key if random generation fails
			// THIS IS NOT SECURE FOR PRODUCTION.
			authKey = "devkey1234567890123456789012345678" // 32 bytes
			log.Println("WARNING: Failed to generate random key for session authentication. Using a hardcoded fallback key. THIS IS NOT SECURE.")
		} else {
			authKey = hex.EncodeToString(key)[:32] // Use 32 bytes of the hex encoded string
		}
		log.Println("WARNING: SESSION_AUTHENTICATION_KEY not set. Using a generated or fallback key for development.")
	}

	// Encryption key (optional, for encrypted cookies if you use NewCookieStore with two keys)
	// For simplicity, we are using only an authentication key which HMACs the cookie but doesn't encrypt its content.
	// If session data is sensitive and needs to be hidden from the user (even if they can't tamper with it),
	// then an encryption key should also be used.
	// encryptionKey := os.Getenv("SESSION_ENCRYPTION_KEY")
	// if encryptionKey == "" {
	// 	key := make([]byte, 32) // Must be 16, 24, or 32 bytes long for AES
	// 	_, err := rand.Read(key)
	// 	if err != nil {
	// 		encryptionKey = "enckey1234567890123456789012345678" // 32 bytes for AES-256
	//      log.Println("WARNING: Failed to generate random key for session encryption. Using a hardcoded fallback key. THIS IS NOT SECURE.")
	// 	} else {
	// 		encryptionKey = hex.EncodeToString(key)[:32]
	// 	}
	// 	log.Println("WARNING: SESSION_ENCRYPTION_KEY not set. Using a generated key for development.")
	// }
	// Store = sessions.NewCookieStore([]byte(authKey), []byte(encryptionKey))

	Store = sessions.NewCookieStore([]byte(authKey))

	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,      // Prevent JavaScript access to the cookie
		Secure:   false,     // Set to true if using HTTPS in production
		SameSite: http.SameSiteLaxMode,
	}
}
