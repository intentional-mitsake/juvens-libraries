package database

import (
	"database/sql"
	"time"

	"juvens-library/internal/config"
	"log/slog"
	"os"

	_ "github.com/lib/pq"
)

var logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

func OpenDB() (*sql.DB, error) {
	dbCnfg := config.LoadDBConfig()
	db, err := sql.Open(dbCnfg.DBDriver, dbCnfg.DBSource)
	if err != nil {
		return nil, err
	}
	err = InitializeDB(db)
	if err != nil {
		return nil, err
	}
	logger.Info("Database initialized")
	p := db.Ping()
	if p != nil {
		return nil, p
	}
	return db, nil
}

func CloseDB(db *sql.DB) error {
	err := db.Close()
	if err != nil {
		return err
	}
	return nil
}

func InitializeDB(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS users (
        id TEXT PRIMARY KEY,
        email TEXT UNIQUE NOT NULL,
		username TEXT NOT NULL
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
		//no need to log since the caller will log the error, just return it here
		return err
	}

	logger.Info("Database schema verified/created successfully")
	return nil
}

func InsertLoginInfo(db *sql.DB, email, name, encAccessToken, encRefreshToken, hashedSessionID string, expiry time.Time) error {
	query := `
		INSERT INTO users (id, email, username) 
		VALUES ($1, $2, $3)
		ON CONFLICT (email) DO NOTHING

		INSERT INTO tokens (user_id, access_token, refresh_token, expiry, session_id)
		VALUES ($1, $4, $5, $6, $7)
		ON CONFLICT (session_id) DO UPDATE 
		SET access_token = EXCLUDED.access_token, refresh_token = EXCLUDED.refresh_token, expiry = EXCLUDED.expiry
	`
	_, err := db.Exec(query, hashedSessionID, email, name, encAccessToken, encRefreshToken, expiry, hashedSessionID)
	if err != nil {
		return err
	}
	return nil
}
