package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
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

	if env == "LOCAL" {
		err := godotenv.Load()
		if err != nil {
			return nil, err
		}
	}

	return &APIConfig{
		Env: env,

		Host:             os.Getenv("HOST"),
		Port:             os.Getenv("PORT"),
		AllowedOrigins:   strings.Split(os.Getenv("ALLOWED_ORIGINS"), ","),
		AllowedMethods:   strings.Split(os.Getenv("ALLOWED_METHODS"), ","),
		AllowedHeaders:   strings.Split(os.Getenv("ALLOWED_HEADERS"), ","),
		ExposedHeaders:   strings.Split(os.Getenv("EXPOSED_HEADERS"), ","),
		AllowCredentials: getEnvAsBool(os.Getenv("ALLOW_CREDENTIALS")),
		MaxAge:           getEnvAsInt(os.Getenv("MAX_AGE")),
	}, nil
}

func getEnvAsBool(str string) bool {
	value, err := strconv.ParseBool(str)

	if err != nil {
		log.Printf("[ERROR] Error parsing env variable %s to bool: %v", str, err)
	}

	return value
}

func getEnvAsInt(str string) int {
	value, err := strconv.Atoi(str)

	if err != nil {
		log.Printf("[ERROR] Error parsing env variable %s to int: %v", str, err)
	}

	return value
}
