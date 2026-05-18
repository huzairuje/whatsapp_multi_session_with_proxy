package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"whatsapp_multi_session/boot"
	"whatsapp_multi_session/config"
	"whatsapp_multi_session/primitive"

	"github.com/gin-gonic/gin"
	"github.com/gookit/event"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	//parse flag
	flag.StringVar(&config.Env, "env", "local", "A config name that used by server")
	flag.Parse()

	//initialize config
	config.Initialize()

	if strings.ToUpper(config.Conf.Env) == strings.ToUpper(config.EnvironmentProd) {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create a new Gin router
	router := gin.Default()

	router.NoRoute(func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Not Matching of Any Routes"})
		return
	})

	router.NoMethod(func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Method Not Allowed"})
		return
	})

	//initiate app
	appRoutes := boot.Setup(router)

	port := fmt.Sprintf(":%v", config.Conf.Port)
	if port == "" {
		port = fmt.Sprintf(":%v", 1234)
	}

	log.Printf("Server running on port %s", port)
	serve := &http.Server{
		Addr:    port,
		Handler: appRoutes,
	}

	//start pprof server
	if config.Conf.Pprof.Enable {
		log.Printf("pprof is enabled")
		go func() {
			portPprof := fmt.Sprintf("%v", config.Conf.Pprof.PprofPort)
			if portPprof == "" {
				portPprof = fmt.Sprintf("%v", 5555)
			}
			log.Printf("pprof running on port :%s", portPprof)
			if err := http.ListenAndServe(fmt.Sprintf("%s:%v", config.Conf.Pprof.PprofAddress, portPprof), nil); err != nil {
				log.Fatal("shutting down the server pprof")
			}
		}()
	}

	// Start server
	go func() {
		if err := serve.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("shutting down the server")
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server.
	quit := make(chan os.Signal, 1) // <- buffered channel of size 1
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutdown Server ...")
	event.MustFire(primitive.ShutDownEvent, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := serve.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	select {
	case <-ctx.Done():
		log.Println("timeout of 1 second.")
	}
	log.Println("Server exiting")
}
