package config

import (
	"os"
)

type RedisConfig struct {
	UseTLS bool

	Host     string
	Port     string
	Password string
}

func NewRedisConfig() (*RedisConfig, error) {
	return &RedisConfig{
		UseTLS:   getEnvAsBool("RD_USE_TLS"),
		Host:     os.Getenv("RD_HOST"),
		Port:     os.Getenv("RD_PORT"),
		Password: os.Getenv("RD_PASSWORD"),
	}, nil
}
