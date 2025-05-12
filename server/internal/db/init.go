package db

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

//go:embed config/schema.sql
var schemaGenSql string

func NewDatabase() (*sql.DB, error) {
	dbPool, err := sql.Open("sqlite", "db.sqlite?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		reason := fmt.Sprintf("error opening database: %v", err)
		log.Fatalln(reason)
		return nil, errors.New(reason)
	}

	log.Println("Initializing database")
	if _, err := dbPool.ExecContext(context.Background(), schemaGenSql); err != nil {
		reason := fmt.Sprintf("error initializing database: %v", err)
		log.Fatalln(reason)
		return nil, errors.New(reason)
	}

	return dbPool, nil
}
