package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jose-lico/go-plate/config"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gLogger "gorm.io/gorm/logger"
)

func NewSQLGormDB(cfg *config.SQLGormConfig, logger *zap.Logger) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.DatabaseName,
		cfg.SSLMode,
		func() string {
			if cfg.SSLMode == "verify-full" {
				return fmt.Sprintf(" sslrootcert=%s", "cert will be here in the future")
			}
			return ""
		}())

	newLogger := gLogger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		gLogger.Config{
			SlowThreshold: time.Second,
			LogLevel:      gLogger.Silent,
			Colorful:      true,
		},
	)

	var db *gorm.DB
	var err error

	for attempts := 0; attempts < maxAttempts; attempts++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: newLogger,
		})
		if err != nil {
			if attempts+1 < maxAttempts {
				logger.Warn(fmt.Sprintf("Failed to connect to Postgres (attempt %d). Attempting again in %v...", attempts+1, reconnectCooldown), zap.Error(err))
			}
			time.Sleep(reconnectCooldown)
		} else {
			logger.Info("Connected to Postgres")
			return db, nil
		}
	}

	return nil, fmt.Errorf("failed to connect to Postgres after %d attempts, error: %w", maxAttempts, err)
}
