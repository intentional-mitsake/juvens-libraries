package database

import (
	"database/sql"
	"juvens-library/internal/config"
	"log/slog"
	"os"
	//"github.com/lib/pq"
)

var logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

func OpenDB() (*sql.DB, error) {
	dbCnfg := config.LoadDBConfig()
	db, err := sql.Open(dbCnfg.DBDriver, dbCnfg.DBSource)
	if err != nil {
		logger.Error("Failed to open database", "error", err)
		return nil, err
	}
	p := db.Ping()
	if p != nil {
		logger.Error("Failed to ping database", "error", p)
		return nil, p
	}
	return db, nil
}

func CloseDB(db *sql.DB) error {
	err := db.Close()
	if err != nil {
		logger.Error("Failed to close database", "error", err)
		return err
	}
	return nil
}

func CreateTables(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS users (
        id TEXT PRIMARY KEY,
        email TEXT UNIQUE NOT NULL
    );

    CREATE TABLE IF NOT EXISTS tokens (
        user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
        access_token TEXT NOT NULL,
        refresh_token TEXT NOT NULL,
        expiry TIMESTAMPTZ NOT NULL,
        session_id TEXT PRIMARY KEY
    );

    CREATE TABLE IF NOT EXISTS books (
        id TEXT PRIMARY KEY,
        title TEXT NOT NULL,
        author TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS user_books (
        user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
        book_id TEXT REFERENCES books(id) ON DELETE CASCADE,
        started_at TIMESTAMPTZ,
        finished_at TIMESTAMPTZ,
        rating INT DEFAULT 0 CHECK (rating >= 0 AND rating <= 5),
        rstatus INT DEFAULT 0,
        PRIMARY KEY (user_id, book_id)
    );
    `

	_, err := db.Exec(query)
	if err != nil {
		logger.Error("Failed to initialize database schema", "error", err)
		return err
	}

	logger.Info("Database schema verified/created successfully")
	return nil
}
