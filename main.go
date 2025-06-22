// main.go
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall" // Required for signal handling

	"whatsapp_multi_session/boot"
	"whatsapp_multi_session/config"
	// "whatsapp_multi_session/database" // No longer directly used here
	// "whatsapp_multi_session/app/db" // No longer directly used here
	// "whatsapp_multi_session/app/handlers" // No longer directly used here
	// "whatsapp_multi_session/app/sessions" // No longer directly used here
	// "github.com/gorilla/mux" // No longer used

	"github.com/gin-gonic/gin"
)

func main() {
	// Load config
	// This assumes config.LoadConfig() is the correct function to initialize config.Conf
	if err := config.LoadConfig(); err != nil {
		 log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Println("Starting application...")

	// Create a new Gin engine
	// gin.Default() includes Logger and Recovery middleware.
	router := gin.Default()

	// Call boot.Setup to initialize database, routes, etc.
	// boot.Setup now handles all previous initializations (DB, sessions, templates, WA routes, dashboard routes)
	// and returns the configured Gin engine.
	configuredRouter := boot.Setup(router)

	// Determine port from config or default
	port := config.Conf.Port
	if port == "" {
		port = "8080" // Default port if not specified in config
		log.Printf("Port not specified in config, using default port: %s", port)
	}

	log.Printf("Server starting on port %s", port)

	// Start the server in a goroutine so it doesn't block the graceful shutdown handling.
	go func() {
		if err := configuredRouter.Run(":" + port); err != nil && err != http.ErrServerClosed {
			// http.ErrServerClosed is expected on graceful shutdown, so don't log it as a fatal error.
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

    // Graceful Shutdown
    quit := make(chan os.Signal, 1)
    // signal.Notify registers the given channel to receive notifications of the specified signals.
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    // Block until a signal is received.
    <-quit
    log.Println("Shutting down server...")

    // Add any cleanup tasks here if needed. For example, closing database connections
    // if they are not managed by a higher-level component or if the application
    // doesn't handle this automatically on exit.
    // The current WA application might have its own shutdown triggers (e.g., listener.TriggerShutdown()).
    // These would need to be integrated here if applicable.

    // For example, if listener package has a global shutdown trigger:
    // listener.TriggerGlobalShutdown() // This is hypothetical

    log.Println("Server gracefully stopped.")
}
