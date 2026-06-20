package config

import (
	"fmt"
	"os"
)

type Config struct {
	ServerAddr  string
	DatabaseDSN string
}

func Load() (*Config, error) {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := getEnv("DB_SSLMODE", "disable")

	if user == "" {
		return nil, fmt.Errorf("DB_USER is required")
	}
	if password == "" {
		return nil, fmt.Errorf("DB_PASSWORD is required")
	}
	if dbname == "" {
		return nil, fmt.Errorf("DB_NAME is required")
	}

	cfg := &Config{
		ServerAddr: getEnv("SERVER_ADDR", ":8080"),
		// pgx понимает и URL, и key=value формат
		DatabaseDSN: fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, password, dbname, sslmode,
		),
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
