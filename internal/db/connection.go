package db

import (
	"database/sql"
	"log"
)

func NewDbConnection(connectionString string) *sql.DB {
	dbConn, err := sql.Open("pgx", connectionString)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}

	if err := dbConn.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	return dbConn
}
