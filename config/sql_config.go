package config

import (
	"os"
)

type SQLGormConfig struct {
	SSLMode     string
	SSLCertPath string

	Host         string
	Port         string
	Username     string
	Password     string
	DatabaseName string
}

func NewSQLConfig() *SQLGormConfig {
	return &SQLGormConfig{
		SSLMode:      os.Getenv("SQL_SSL_MODE"),
		SSLCertPath:  os.Getenv("SQL_SSL_CERT_PATH"),
		Host:         os.Getenv("SQL_HOST"),
		Port:         os.Getenv("SQL_PORT"),
		Username:     os.Getenv("SQL_USER"),
		Password:     os.Getenv("SQL_PASSWORD"),
		DatabaseName: os.Getenv("SQL_NAME"),
	}
}
