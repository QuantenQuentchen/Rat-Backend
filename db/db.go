package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var db *sqlx.DB

func connectDB() error {
	dbFile := "mydb.sqlite"
	var err error
	db, err = sqlx.Connect("sqlite3", dbFile)
	return err
}

func InitDB() error {
	// This will create or open the file-based SQLite DB
	err := connectDB()
	if err != nil {
		return fmt.Errorf("failed to connect to SQLite: %w", err)
	}
	err = createSchemas()
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}
	err = createRoles()
	if err != nil {
		return fmt.Errorf("failed to create Roles: %w", err)
	}
	return nil
}
