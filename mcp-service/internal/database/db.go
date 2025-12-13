package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

// DB wraps sql.DB with additional functionality
type DB struct {
	*sql.DB
}

// Connect creates a new database connection
func Connect(connectionString string) (*DB, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✅ Database connected successfully")

	database := &DB{db}

	// Initialize schema
	if err := database.InitSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return database, nil
}

// InitSchema runs the database schema
func (db *DB) InitSchema() error {
	schemaPath := "database/schema.sql"
	
	// Read schema file
	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		// Try alternative path
		schema, err = os.ReadFile("../database/schema.sql")
		if err != nil {
			log.Printf("Warning: Could not read schema file: %v", err)
			return nil // Don't fail if schema file not found (might be in production)
		}
	}

	// Execute schema
	_, err = db.Exec(string(schema))
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	log.Println("✅ Database schema initialized")
	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}
