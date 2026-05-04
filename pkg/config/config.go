package config

import "os"

type Config struct {
	Port        string
	DatabaseURL string
	Env         string
	JWTSecret   string
	FrontendURL string
}

func Load() Config {
	return Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgresql://neondb_owner:npg_ReLYa72BIrDn@ep-plain-union-al4ogt7p-pooler.c-3.eu-central-1.aws.neon.tech/neondb?sslmode=require&channel_binding=require"),
		Env:         getEnv("ENV", "development"),
		JWTSecret:   getEnv("JWT_SECRET", "change-me-in-production"),
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
