package boot

import (
	"database/sql"
	"fmt"

	"whatsapp_multi_session/auth"
	"whatsapp_multi_session/commandhandler"
	"whatsapp_multi_session/config"
	"whatsapp_multi_session/cronjob"
	"whatsapp_multi_session/database"
	"whatsapp_multi_session/handler"
	"whatsapp_multi_session/listener"
	"whatsapp_multi_session/proxy"
	"whatsapp_multi_session/routers"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow/store/sqlstore"
)

func Setup(r *gin.Engine) *gin.Engine {
	//initiate database sqlite
	var connDb *sqlstore.Container
	var rawDB interface{}
	var isPostgres bool
	
	if config.Conf.Postgres.EnablePostgres {
		postgresConn, err := database.NewPostgresClient(&config.Conf)
		if err != nil {
			log.Errorf("error from initiate postgresql : %v ", err)
			panic(err)
		}
		connDb = postgresConn
		
		rawDB, err = database.GetRawPostgresDB(&config.Conf)
		if err != nil {
			log.Errorf("error getting raw postgres db: %v", err)
			panic(err)
		}
		isPostgres = true
	} else {
		sqliteConn, err := database.NewSqlite()
		if err != nil {
			log.Errorf("error from initiate sqlite : %v ", err)
			panic(err)
		}
		connDb = sqliteConn
		
		rawDB, err = database.GetRawSqliteDB()
		if err != nil {
			log.Errorf("error getting raw sqlite db: %v", err)
			panic(err)
		}
		isPostgres = false
	}

	// Initialize auth service
	authService := auth.NewService(rawDB.(*sql.DB), "your-secret-key-change-this", isPostgres)
	if err := authService.InitializeDatabase(); err != nil {
		log.Errorf("error initializing auth database: %v", err)
		panic(err)
	}

	// load proxy list
	proxyManager := proxy.NewManager()
	err := proxyManager.LoadFromFile(config.Conf.Proxy.Directory)
	if err != nil {
		log.Fatal("Error loading proxy list:", err)
	}

	//initiate command handler here
	cmdHandler := commandhandler.NewCommandHandler(connDb, proxyManager)

	listen := listener.NewListener(cmdHandler)
	go func() {
		// listener on trigger start up
		listen.TriggerStartUp()
	}()
	//listener on trigger shutdown
	listen.ListenForShutdownEvent()

	newHandler := handler.NewHandler(cmdHandler, authService)
	router := routers.NewRoutes(newHandler, authService)
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
