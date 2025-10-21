package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/zaeem.arshad/rmq-stream-viewer/internal/api"
	"github.com/zaeem.arshad/rmq-stream-viewer/internal/config"
	"github.com/zaeem.arshad/rmq-stream-viewer/internal/rabbitmq"
)

//go:embed static
var staticFiles embed.FS

func main() {
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Loaded configuration with %d connection(s)", len(cfg.Connections))

	// Create RabbitMQ manager
	manager := rabbitmq.NewManager(cfg.Connections)

	// Connect to RabbitMQ instances
	ctx := context.Background()
	if err := manager.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer manager.Close()

	log.Println("Successfully connected to all RabbitMQ instances")

	// Create HTTP handler
	handler := api.NewHandler(manager)

	// Setup router
	router := mux.NewRouter()

	// Register API routes
	handler.RegisterRoutes(router)

	// Serve static files (React frontend)
	// Try serving from web/dist for development
	if _, err := os.Stat("web/dist"); err == nil {
		router.PathPrefix("/").Handler(http.FileServer(http.Dir("web/dist")))
		log.Println("Serving frontend from web/dist")
	} else {
		// Try embedded files
		staticFS, err := fs.Sub(staticFiles, "static")
		if err == nil {
			router.PathPrefix("/").Handler(http.FileServer(http.FS(staticFS)))
			log.Println("Serving embedded frontend")
		} else {
			log.Println("Warning: No frontend files found. API-only mode.")
		}
	}

	// Apply middleware
	router.Use(api.LoggingMiddleware)
	router.Use(api.CORSMiddleware)

	// Create server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}

