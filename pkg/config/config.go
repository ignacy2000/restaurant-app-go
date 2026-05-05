package config

import "os"

type Config struct {
	Port             string
	DatabaseURL      string
	Env              string
	JWTSecret        string
	FrontendURL      string
	RedisAddr        string
	GmailFrom        string
	GmailAppPassword string
}

func Load() Config {
	return Config{
		Port:             getEnv("PORT", "8080"),
		DatabaseURL:      getEnv("DATABASE_URL", "postgresql://************************"),
		Env:              getEnv("ENV", "development"),
		JWTSecret:        getEnv("JWT_SECRET", "change-me-in-production"),
		FrontendURL:      getEnv("FRONTEND_URL", "http://localhost:3000"),
		RedisAddr:        getEnv("REDIS_ADDR", "localhost:6379"),
		GmailFrom:        getEnv("GMAIL_FROM", "email@gmail.com"),
		GmailAppPassword: getEnv("GMAIL_APP_PASSWORD", "***********"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
