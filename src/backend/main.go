package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Eduardo-Barreto/web-ponderada/backend/auth"
	"github.com/Eduardo-Barreto/web-ponderada/backend/config"
	"github.com/Eduardo-Barreto/web-ponderada/backend/database"
	"github.com/Eduardo-Barreto/web-ponderada/backend/repository"
	"github.com/Eduardo-Barreto/web-ponderada/backend/routes"
	"github.com/Eduardo-Barreto/web-ponderada/backend/storage"
)

func main() {
	// 1. Load Configuration
	config.LoadConfig()

	// 2. Initialize Authentication (JWT Secret)
	auth.InitializeAuth()

	// 3. Connect to Database
	database.ConnectDB()
	defer database.CloseDB() // Ensure pool is closed on exit

	// 4. Initialize Repositories
	userRepo := repository.NewPostgresUserRepository(database.Pool)
	productRepo := repository.NewPostgresProductRepository(database.Pool)
	// Use local storage implementation
	fileRepo := storage.NewLocalStorage() // Create local storage instance

	// 5. Setup Router
	router := routes.SetupRouter(userRepo, productRepo, fileRepo) // Pass fileRepo

	// 6. Start Server with Graceful Shutdown
	server := &http.Server{
		Addr:    ":" + config.AppConfig.Port,
		Handler: router,
		// Optional: Add timeouts for security and stability
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Run server in a goroutine so it doesn't block
	go func() {
		log.Printf("Server starting on port %s", config.AppConfig.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s: %v\n", config.AppConfig.Port, err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can't be caught, so don't need it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the requests it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
