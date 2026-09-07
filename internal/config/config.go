package config

import (
	"fmt"
	"net"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	JWTSecret      string
	AllowedOrigins string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:           os.Getenv("PORT"),
		DBHost:         os.Getenv("DB_HOST"),
		DBPort:         os.Getenv("DB_PORT"),
		DBUser:         os.Getenv("DB_USER"),
		DBPassword:     os.Getenv("DB_PASSWORD"),
		DBName:         os.Getenv("DB_NAME"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		AllowedOrigins: envOrDefault("ALLOWED_ORIGINS", "http://localhost:5173"),
	}

	if cfg.DBHost == "" || cfg.DBUser == "" || cfg.DBName == "" || cfg.JWTSecret == "" {
		return nil, fmt.Errorf("faltan variables de entorno requeridas para la base de datos o autenticacion")
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func (c *Config) DSN() string {
	dsn := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.DBUser, c.DBPassword),
		Host:     net.JoinHostPort(c.DBHost, c.DBPort),
		Path:     c.DBName,
		RawQuery: "sslmode=disable",
	}
	return dsn.String()
}
