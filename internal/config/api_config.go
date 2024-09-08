package config

import (
	"os"
	"strings"
)

type APIConfig struct {
	Env string

	Host string
	Port string

	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

func NewAPIConfig() (*APIConfig, error) {
	env := os.Getenv("ENV")

	return &APIConfig{
		Env: env,

		Host:             os.Getenv("HOST"),
		Port:             os.Getenv("PORT"),
		AllowedOrigins:   strings.Split(os.Getenv("ALLOWED_ORIGINS"), ","),
		AllowedMethods:   strings.Split(os.Getenv("ALLOWED_METHODS"), ","),
		AllowedHeaders:   strings.Split(os.Getenv("ALLOWED_HEADERS"), ","),
		ExposedHeaders:   strings.Split(os.Getenv("EXPOSED_HEADERS"), ","),
		AllowCredentials: getEnvAsBool("ALLOW_CREDENTIALS"),
		MaxAge:           getEnvAsInt("MAX_AGE"),
	}, nil
}
