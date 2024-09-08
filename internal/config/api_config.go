package config

import (
	"os"

	"github.com/joho/godotenv"
)

type APIConfig struct {
	Env string

	Host string
	Port string
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

		Host: os.Getenv("HOST"),
		Port: os.Getenv("PORT"),
	}, nil
}
