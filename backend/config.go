package main

import "os"

type Config struct {
	Port      string
	DBDSN     string
	JWTSecret string
}

func LoadConfig() Config {
	return Config{
		Port:      getenv("PORT", "8010"),
		DBDSN:     getenv("DB_DSN", "root:knowbase_dev@tcp(127.0.0.1:3306)/yunyue?charset=utf8mb4&parseTime=True"),
		JWTSecret: getenv("JWT_SECRET", "dev-secret-change-me"),
	}
}

func getenv(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}
