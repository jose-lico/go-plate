package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jose-lico/go-plate/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewSQLGormDB(cfg *config.SQLGormConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.DatabaseName,
		cfg.SSLMode,
		func() string {
			if cfg.SSLMode == "verify-full" {
				return fmt.Sprintf(" sslrootcert=%s", "cert will be here in the future")
			}
			return ""
		}())

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Silent,
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
				log.Printf("[ERROR] Failed to connect to Postgres (attempt %d). Error: %v. Attempting again in %v...", attempts+1, err, reconnectCooldown)
			} else {
				log.Printf("[ERROR] Failed to connect to Postgres (attempt %d). Error: %v", attempts+1, err)
			}
			time.Sleep(reconnectCooldown)
		} else {
			log.Println("[TRACE] Connected to Postgres")
			return db, nil
		}
	}

	return nil, fmt.Errorf("failed to connect to Postgres after %d attempts, error: %w", maxAttempts, err)
}
