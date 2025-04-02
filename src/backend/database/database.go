package database

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Eduardo-Barreto/web-ponderada/backend/config"
)

var Pool *pgxpool.Pool

func ConnectDB() {
	var err error
	// Use AppConfig loaded earlier
	dbUrl := config.AppConfig.DatabaseURL

	// pgxpool configuration
	dbConfig, err := pgxpool.ParseConfig(dbUrl)
	if err != nil {
		log.Fatalf("Unable to parse database URL: %v\n", err)
	}

	dbConfig.MaxConns = 10
	dbConfig.MinConns = 2
	dbConfig.MaxConnLifetime = time.Hour
	dbConfig.MaxConnIdleTime = 30 * time.Minute
	dbConfig.HealthCheckPeriod = time.Minute
	dbConfig.ConnConfig.ConnectTimeout = 5 * time.Second

	// Retry connection logic
	maxRetries := 5
	retryDelay := 5 * time.Second

	for i := 1; i <= maxRetries; i++ {
		Pool, err = pgxpool.NewWithConfig(context.Background(), dbConfig)
		if err != nil {
			log.Printf("Attempt %d: Unable to create connection pool: %v. Retrying in %v...", i, err, retryDelay)
			time.Sleep(retryDelay)
			continue
		}

		// Test the connection
		err = Pool.Ping(context.Background())
		if err == nil {
			log.Println("Successfully connected to PostgreSQL!")
			return
		}
	}

	log.Fatalf("Unable to connect to database after %d attempts: %v\n", maxRetries, err)
	os.Exit(1)
}

// Optional: Function to close the pool gracefully on shutdown
func CloseDB() {
	if Pool != nil {
		log.Println("Closing database connection pool...")
		Pool.Close()
	}
}
