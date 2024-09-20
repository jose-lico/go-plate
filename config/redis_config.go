package config

import (
	"os"

	"github.com/jose-lico/go-plate/utils"
)

type RedisConfig struct {
	UseTLS bool

	Host     string
	Port     string
	Password string
}

func NewRedisConfig() (*RedisConfig, error) {
	return &RedisConfig{
		UseTLS:   utils.GetEnvAsBool("RD_USE_TLS"),
		Host:     os.Getenv("RD_HOST"),
		Port:     os.Getenv("RD_PORT"),
		Password: os.Getenv("RD_PASSWORD"),
	}, nil
}
