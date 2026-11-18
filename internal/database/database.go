package database

import (
	"database/sql"
	_ "embed"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

//go:embed schema.sql
var schema string

type DB struct {
	conn *sql.DB
}

// Connect établit la connexion à PostgreSQL (avec quelques essais : la base peut démarrer après nous)
func Connect(connStr string) (*DB, error) {
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	for attempt := 1; ; attempt++ {
		err = conn.Ping()
		if err == nil {
			break
		}
		if attempt >= 10 {
			return nil, fmt.Errorf("failed to ping database: %w", err)
		}
		time.Sleep(2 * time.Second)
	}

	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)

	return &DB{conn: conn}, nil
}

// Migrate crée ou met à jour le schéma (idempotent)
func (db *DB) Migrate() error {
	if _, err := db.conn.Exec(schema); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	return nil
}

// Close ferme la connexion
func (db *DB) Close() error {
	return db.conn.Close()
}
