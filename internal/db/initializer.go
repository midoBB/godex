package db

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/glebarez/sqlite"
	gap "github.com/muesli/go-app-paths"
	"gorm.io/gorm"
)

//go:embed schema.sql
var schema string

func dbExists() (bool, error) {
	scope := gap.NewScope(gap.User, "godex")
	dbPath, err := scope.DataPath("godex.sqlite")
	if err != nil {
		return false, err
	}
	_, err = os.Stat(dbPath)
	return err == nil, nil
}

func createDB(dbPath string) error {
	// Open the SQLite database
	db, err := gorm.Open(sqlite.Open(dbPath))
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Create the database schema
	result := db.Exec(schema)
	if result.Error != nil {
		return fmt.Errorf("failed to execute SQL script: %w", result.Error)
	}

	return nil
}
