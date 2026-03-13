package db

import (
	"database/sql"
	"log"

	"github.com/pressly/goose/v3"
)

func MigrateDB(db *sql.DB) {
	migrationsDir := "./internal/db/migrations"

	if err := goose.Up(db, migrationsDir); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}
	log.Println("Database migrated successfully!")
}
