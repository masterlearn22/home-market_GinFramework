package config

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func SetupServer(app *gin.Engine) {
	// 1. Baca port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 2. Definisikan HTTP server
	server := &http.Server{
		Addr:    ":" + port,
		Handler: app,
	}

	// 3. Jalankan server goroutine
	go func() {
		log.Printf("Server running on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server stopped: %v", err)
		}
	}()

	// 4. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
