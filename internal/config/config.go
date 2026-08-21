package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config holds all environment-driven settings for the application.
type Config struct {
	AppPort string
	AppEnv  string

	DatabaseURL string

	StorageEndpoint  string
	StorageBucket    string
	StorageAccessKey string
	StorageSecretKey string

	OpenAIAPIKey   string
	OpenAIBaseURL  string
	OpenAIModel    string
	OpenAISTTModel string
	AIMockImages   bool

	JWTSecret string
}

// Load reads .env (if present) and environment variables into a Config.
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppPort: getEnv("APP_PORT", "8080"),
		AppEnv:  getEnv("APP_ENV", "development"),

		DatabaseURL: getEnv("DATABASE_URL", ""),

		StorageEndpoint:  getEnv("STORAGE_ENDPOINT", ""),
		StorageBucket:    getEnv("STORAGE_BUCKET", ""),
		StorageAccessKey: getEnv("STORAGE_ACCESS_KEY", ""),
		StorageSecretKey: getEnv("STORAGE_SECRET_KEY", ""),

		OpenAIAPIKey:   getEnv("OPENAI_API_KEY", ""),
		OpenAIBaseURL:  getEnv("OPENAI_BASE_URL", ""),
		OpenAIModel:    getEnv("OPENAI_MODEL", "llama-3.3-70b-versatile"),
		OpenAISTTModel: getEnv("OPENAI_STT_MODEL", "whisper-large-v3"),
		AIMockImages:   getEnv("AI_MOCK_IMAGES", "true") == "true",

		JWTSecret: getEnv("JWT_SECRET", "dev-insecure-secret-change-me"),
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
