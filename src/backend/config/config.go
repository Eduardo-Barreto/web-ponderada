package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	Port        string
	JWTSecret   string
	UploadDir   string
}

var AppConfig *Config

func LoadConfig() {
	godotenv.Load()

	AppConfig = &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://user:password@db:5432/webappdb?sslmode=disable"),
		Port:        getEnv("PORT", "8080"),
		JWTSecret:   getEnv("JWT_SECRET", "a-very-secret-key"),
		UploadDir:   getEnv("UPLOAD_DIR", "./uploads"), // Relative to backend executable or volume mount
	}

	// Ensure upload directory exists
	if _, err := os.Stat(AppConfig.UploadDir); os.IsNotExist(err) {
		log.Printf("Upload directory %s does not exist, creating...", AppConfig.UploadDir)
		err = os.MkdirAll(AppConfig.UploadDir, 0755) // Read/write for user, read/execute for others
		if err != nil {
			log.Fatalf("Failed to create upload directory: %v", err)
		}
	} else {
		log.Printf("Using upload directory: %s", AppConfig.UploadDir)
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	log.Printf("Environment variable %s not set, using fallback: %s", key, fallback)
	return fallback
}

// Helper to get env var as int
func getEnvAsInt(key string, fallback int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	log.Printf("Environment variable %s not a valid integer, using fallback: %d", key, fallback)
	return fallback
}
