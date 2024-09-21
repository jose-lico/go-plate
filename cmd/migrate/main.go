package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jose-lico/go-plate/config"
	"github.com/jose-lico/go-plate/utils"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Please specify 'up' or 'down' as a command argument and path to migrations.")
		return
	}

	direction := os.Args[1]
	migrationsPath := os.Args[2]

	env := os.Getenv("ENV")

	if env == "LOCAL" {
		err := utils.LoadEnvs()
		if err != nil {
			log.Fatalf("[FATAL] Error loading .env: %v", err)
		}
	}

	cfg, err := config.NewSQLConfig()

	if err != nil {
		log.Fatalf("Error loading SQL Config: %v", err)
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.DatabaseName,
		cfg.SSLMode,
		func() string {
			if cfg.SSLMode == "verify-full" {
				return fmt.Sprintf(" sslrootcert=%s", "cert will be here in the future")
			}
			return ""
		}())

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Could not connect to the database: %v\n", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		log.Fatalf("Could not create database driver: %v\n", err)
	}

	migrationPath := "file://" + migrationsPath

	m, err := migrate.NewWithDatabaseInstance(migrationPath, "postgres", driver)
	if err != nil {
		log.Fatalf("Could not create migrate instance: %v\n", err)
	}

	switch direction {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration up failed: %v\n", err)
		} else {
			fmt.Println("Migration up successful or no changes to apply.")
		}
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration down failed: %v\n", err)
		} else {
			fmt.Println("Migration down successful.")
		}
	default:
		fmt.Println("Invalid argument. Please specify 'up' or 'down'.")
	}
}
