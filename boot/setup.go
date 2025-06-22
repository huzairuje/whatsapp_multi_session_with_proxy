// boot/setup.go
package boot

import (
	"fmt"
	"os"
    "log" // Standard log package

	"whatsapp_multi_session/commandhandler"
	"whatsapp_multi_session/config"
	"whatsapp_multi_session/cronjob"
	core_database "whatsapp_multi_session/database" // WA app's DB

	dashboard_db "whatsapp_multi_session/app/db"
	"whatsapp_multi_session/app/handlers" // For Init...Templates functions
	dashboard_sessions "whatsapp_multi_session/app/sessions"

	wa_handler "whatsapp_multi_session/handler"
	"whatsapp_multi_session/listener"
	"whatsapp_multi_session/proxy"
	wa_routers "whatsapp_multi_session/routers" // WA app's routers

	"github.com/gin-gonic/gin"
    // Using standard log, remove sirupsen/logrus if not used elsewhere in this file
    // log_logrus "github.com/sirupsen/logrus" // Aliased to avoid conflict if std log is also "log"
	"go.mau.fi/whatsmeow/store/sqlstore"
)

func Setup(r *gin.Engine) *gin.Engine {
    // Using standard log for messages in this function for consistency
	var connDb *sqlstore.Container // WA App DB
	if config.Conf.Postgres.EnablePostgres {
		postgresConn, errPg := core_database.NewPostgresClient(&config.Conf)
		if errPg != nil {
			log.Printf("ERROR from initiate postgresql : %v ", errPg)
			panic(errPg)
		}
		connDb = postgresConn
	} else {
		sqliteConn, errSq := core_database.NewSqlite()
		if errSq != nil {
			log.Printf("ERROR from initiate sqlite : %v ", errSq)
			panic(errSq)
		}
		connDb = sqliteConn
	}

    // ---- START DASHBOARD INITIALIZATION ----
    log.Println("Initializing dashboard components...")
    dashboard_db.ConnectDashboardDB()
    dashboard_db.MigrateDashboardSchema()

    // Session key handling for dashboard_sessions.InitSessionStore()
    // dashboard_sessions.InitSessionStore() internally checks os.Getenv("SESSION_AUTHENTICATION_KEY")
    // We can log a warning here if it's not set, similar to how InitSessionStore does.
    if os.Getenv("SESSION_AUTHENTICATION_KEY") == "" {
        log.Println("WARNING: boot/setup.go: SESSION_AUTHENTICATION_KEY for dashboard is not set. dashboard_sessions.InitSessionStore() will use its default/generated key logic.")
    }
    dashboard_sessions.InitSessionStore()

    // Load HTML templates for Gin
    // The path is relative to the executable's working directory.
    // If templates are in 'app/templates', and executable is in project root, this is correct.
    // For `LoadHTMLGlob`, if templates are directly in `app/templates` (not subdirs), use `app/templates/*.html`
    // If there are templates in subdirectories of `app/templates` that need to be loaded by their relative path from `app/templates`
    // e.g. `c.HTML(..., "subdir/template.html", ...)` then a pattern like `app/templates/**/*` might be needed,
    // but this depends on how Gin handles nested template names.
    // Given current structure, `app/templates/*` should cover all .html files directly in that folder.
    r.LoadHTMLGlob("app/templates/*.html") // Ensure this pattern correctly captures all needed templates.
                                        // If _messages.html is not caught, it might need specific loading or a broader pattern.
                                        // Let's assume for now it's caught.
    log.Println("Dashboard HTML templates loaded for Gin using app/templates/*.html.")

    // Initialize handler-specific templates.
    // These parse templates using html/template, potentially for FuncMaps or if not using Gin's c.HTML directly.
    // Their internal paths (e.g., "../templates/layout.html") are for `go test` context.
    // When run from `main.go` (project root CWD), they would need "app/templates/layout.html".
    // This path discrepancy needs a robust solution (e.g., build tags, config, embed).
    // For now, we acknowledge this will likely fail if `main.go` calls these directly without path adjustment.
    // However, the primary goal here is that Gin loads templates via LoadHTMLGlob for its rendering.
    // These calls might become less important if all rendering shifts to c.HTML("template_name.html", data).
    handlers.InitAuthTemplates()
    handlers.InitDashboardTemplates()
    handlers.InitDeviceTemplates()
    handlers.InitMessageTemplates()
    handlers.InitSettingsTemplates()
    log.Println("Dashboard components (templates, sessions, DB) initialized.")
    // ---- END DASHBOARD INITIALIZATION ----

	proxyManager := proxy.NewManager()
	errProxy := proxyManager.LoadFromFile(config.Conf.Proxy.Directory)
    if errProxy != nil {
        // Using log.Fatalf will exit the program. Consider if just logging an error is preferred.
        log.Fatalf("Error loading proxy list: %v", errProxy)
    }

	cmdHandler := commandhandler.NewCommandHandler(connDb, proxyManager)

    listen := listener.NewListener(cmdHandler)
    go func() {
        listen.TriggerStartUp()
    }()
    listen.ListenForShutdownEvent()

    newWAHandler := wa_handler.NewHandler(cmdHandler)

    routerWA := wa_routers.NewRoutes(newWAHandler)
    r = routerWA.V1(r) // WA routes added to 'r'

    // Add Dashboard Routes
    log.Println("Adding dashboard routes...")
    wa_routers.AddDashboardRoutes(r) // Call the new function to add dashboard routes
                                    // Ensure wa_routers has access to dashboard_handlers

    // Setup static file serving for dashboard (if not already handled globally)
    // Path is relative to executable. If executable is in project root, "./app/static" is correct.
    r.Static("/static", "./app/static")
    log.Println("Static file server for /static mapped to ./app/static")


    cronJobs := cronjob.NewCronJobs(cmdHandler)
    go func() {
        cronJobs.Run()
        // Using standard log for consistency, or ensure log_logrus is properly configured if used.
        log.Println("cronJobs.Run() is successfully initiated")
    }()

	return r
}
