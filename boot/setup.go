package boot

import (
	"fmt"

	"whatsapp_multi_session/commandhandler"
	"whatsapp_multi_session/config"
	"whatsapp_multi_session/cronjob"
	"whatsapp_multi_session/database"
	"whatsapp_multi_session/handler"
	"whatsapp_multi_session/listener"
	"whatsapp_multi_session/proxy"
	"whatsapp_multi_session/routers"
	// Import for dashboard migrations
	dashboardDB "whatsapp_multi_session/dashboard/database"
	dashboardHandlers "whatsapp_multi_session/dashboard/handlers"
	dashboardMiddleware "whatsapp_multi_session/dashboard/middleware"
	dashboardRoutes "whatsapp_multi_session/dashboard/routes"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus" // Already present, but ensure it's used for new logs
	"go.mau.fi/whatsmeow/store/sqlstore"
	"gorm.io/driver/postgres" // For GORM handler initialization
	"gorm.io/driver/sqlite"   // For GORM handler initialization
	"gorm.io/gorm"            // For GORM handler initialization
)

func Setup(r *gin.Engine) *gin.Engine {
	//initiate database
	var connDb *sqlstore.Container
	var dialect string
	var err error

	if config.Conf.Postgres.EnablePostgres {
		log.Info("Initializing PostgreSQL database...")
		postgresConn, errPg := database.NewPostgresClient(&config.Conf)
		if errPg != nil {
			log.Errorf("Error initializing PostgreSQL: %v", errPg)
			panic(errPg)
		}
		connDb = postgresConn
		dialect = "postgres"
		log.Info("PostgreSQL database initialized successfully.")
	} else {
		log.Info("Initializing SQLite database...")
		sqliteConn, errSqlite := database.NewSqlite()
		if errSqlite != nil {
			log.Errorf("Error initializing SQLite: %v", errSqlite)
			panic(errSqlite)
		}
		connDb = sqliteConn
		dialect = "sqlite"
		log.Info("SQLite database initialized successfully.")
	}

	// Run dashboard migrations
	if connDb == nil || connDb.DB == nil {
		log.Fatal("Database connection (connDb or connDb.DB) is nil, cannot run dashboard migrations.")
		// panic("Database connection (connDb or connDb.DB) is nil, cannot run dashboard migrations.")
	}
	log.Info("Running dashboard migrations...")
	if errMig := dashboardDB.RunDashboardMigrations(connDb.DB, dialect); errMig != nil {
		log.Fatalf("Failed to run dashboard migrations: %v", errMig)
		// panic(fmt.Sprintf("Failed to run dashboard migrations: %v", errMig))
	}
	log.Info("Dashboard migrations completed successfully.")

	// Initialize GORM DB instance for dashboard handlers
	var gormForHandlers *gorm.DB
	if dialect == "postgres" {
		gormForHandlers, err = gorm.Open(postgres.New(postgres.Config{
			Conn: connDb.DB,
		}), &gorm.Config{})
	} else if dialect == "sqlite" {
		gormForHandlers, err = gorm.Open(sqlite.Open(connDb.DB.DriverName()+":memory?_foreign_keys=on"), &gorm.Config{})
		// Note: For sqlite, if connDb.DB is from sqlstore.New with a DSN like "file:examplestore.db?_foreign_keys=on",
		// we need to ensure gorm uses the same underlying database.
		// The above sqlite.Open might create an in-memory DB if DSN isn't correctly passed or if connDb.DB.DriverName() is not suitable.
		// A safer way for SQLite if already opened by sqlstore:
		// gormForHandlers, err = gorm.Open(sqlite.Dialector{Conn: connDb.DB}, &gorm.Config{})
		// Re-evaluating the sqlite GORM initialization based on how sqlstore.Container.DB is set up.
		// Assuming connDb.DB is a live connection, Dialector{Conn: connDb.DB} is indeed better.
		gormForHandlers, err = gorm.Open(sqlite.Dialector{Conn: connDb.DB}, &gorm.Config{})
	} else {
		log.Fatalf("Unsupported dialect for GORM handler initialization: %s", dialect)
	}
	if err != nil {
		log.Fatalf("Failed to initialize GORM for dashboard handlers: %v", err)
	}
	log.Info("GORM initialized successfully for dashboard handlers.")

	// load proxy list
	proxyManager := proxy.NewManager()
	err = proxyManager.LoadFromFile(config.Conf.Proxy.Directory)
	if err != nil {
		log.Fatal("Error loading proxy list:", err)
	}

	//initiate command handler here
	cmdHandler := commandhandler.NewCommandHandler(connDb, proxyManager)

	// --- Initialize Dashboard Handlers ---
	log.Info("Initializing dashboard handlers...")
	authHandler := dashboardHandlers.NewAuthHandler(gormForHandlers, cmdHandler, config.Conf.Dashboard.JwtSecretKey, config.Conf.Dashboard.OtpSenderJID)
	deviceHandler := dashboardHandlers.NewDeviceHandler(gormForHandlers, cmdHandler)
	messageHandler := dashboardHandlers.NewMessageHandler(gormForHandlers, cmdHandler)

	// Validate essential dashboard configurations
	if config.Conf.Dashboard.JwtSecretKey == "" {
		log.Fatalf("Dashboard JWT Secret Key is not configured. Please set DASHBOARD_JWT_SECRET_KEY.")
	}
	if config.Conf.Dashboard.OtpSenderJID == "" {
		log.Fatalf("Dashboard OTP Sender JID is not configured. Please set DASHBOARD_OTP_SENDER_JID.")
	}
	log.Info("Dashboard handlers initialized successfully.")

	// --- Setup Dashboard Routes ---
	log.Info("Setting up dashboard routes...")
	dashboardAPIGroup := r.Group("/dashboard")

	// Public Auth Routes
	authPublicRoutes := dashboardAPIGroup.Group("/auth")
	{
		authPublicRoutes.POST("/register", authHandler.Register)
		authPublicRoutes.POST("/request-otp", authHandler.RequestOTP)
		authPublicRoutes.POST("/login", authHandler.Login)
	}

	// Protected Dashboard Routes - Common group with JWT middleware
	protectedDashboardRoutes := dashboardAPIGroup.Group("/") // Will apply to /dashboard/{subpath}
	protectedDashboardRoutes.Use(dashboardMiddleware.JWTMiddleware(config.Conf.Dashboard.JwtSecretKey))
	{
		// Protected Auth Routes
		authProtectedRoutes := protectedDashboardRoutes.Group("/auth")
		{
			authProtectedRoutes.GET("/detail", authHandler.GetAuthDetail)
			authProtectedRoutes.POST("/logout", authHandler.Logout)
		}

		// Protected Device Routes
		// SetupDeviceRoutes registers POST /devices and GET /devices
		// So, they will be /dashboard/devices
		dashboardRoutes.SetupDeviceRoutes(protectedDashboardRoutes.Group("/devices"), deviceHandler)

		// Protected Message & Report Routes
		// SetupMessageRoutes registers /messages and /report sub-groups under protectedDashboardRoutes
		// So, they will be /dashboard/messages/* and /dashboard/report/*
		dashboardRoutes.SetupMessageRoutes(protectedDashboardRoutes, messageHandler)
	}
	log.Info("Dashboard routes set up successfully.")

	listen := listener.NewListener(cmdHandler)
	go func() {
		// listener on trigger start up
		listen.TriggerStartUp()
	}()
	//listener on trigger shutdown
	listen.ListenForShutdownEvent()

	newHandler := handler.NewHandler(cmdHandler)
	router := routers.NewRoutes(newHandler)
	appRoutes := router.V1(r)

	//initiate cronjob
	cronJobs := cronjob.NewCronJobs(cmdHandler)
	go func() {
		// listener on trigger start up
		cronJobs.Run()
		fmt.Println("cronJobs.Run() is successfully initiated")
	}()

	return appRoutes
}
