package database

import (
	"database/sql"
	"time"

	"juvens-library/internal/config"
	"juvens-library/internal/services"
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
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        email TEXT UNIQUE NOT NULL,
		username TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS tokens (
        user_id UUID REFERENCES users(id) ON DELETE CASCADE,
        access_token TEXT NOT NULL,
        refresh_token TEXT NOT NULL,
        expiry TIMESTAMPTZ NOT NULL,
        session_id TEXT PRIMARY KEY
    );

    CREATE TABLE IF NOT EXISTS books (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        title TEXT NOT NULL,
        author TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS user_books (
        user_id UUID REFERENCES users(id) ON DELETE CASCADE,
        book_id UUID REFERENCES books(id) ON DELETE CASCADE,
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
	tx, err := db.Begin()
	// requries ACID compliance as we need to update multi tables
	if err != nil {
		return err
	}
	// insert info, if email already exists, DO NOTIHGN
	// but if use DO Nothng, we cant return the id using REturning, so we use do update but since the email already exists no chagne happens
	userquery := `
		INSERT INTO users (email, username) 
		VALUES ($1, $2)
		ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email
        RETURNING id;
		`
	// user id is gen default randmonly, but if we dont iinsert the user id in the tokens table it will be null
	tokenquery := `
		INSERT INTO tokens (user_id, access_token, refresh_token, expiry, session_id)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (session_id) DO UPDATE 
		SET access_token = EXCLUDED.access_token, refresh_token = EXCLUDED.refresh_token, expiry = EXCLUDED.expiry;
	`
	var generatedUserID string                                       // to store the new or exsitng id
	err = tx.QueryRow(userquery, email, name).Scan(&generatedUserID) // queeryrow as we need to get the generated id
	if err != nil {
		tx.Rollback() // if there is an error in the transaction, we need to rollback to maintain data integrity
		return err
	}
	_, err = tx.Exec(tokenquery, generatedUserID, encAccessToken, encRefreshToken, expiry, hashedSessionID)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func ValidateSessionID(db *sql.DB, sessionID string) (bool, error) {
	hashedSessionID, err := services.HashSessionID(sessionID)
	if err != nil {
		return false, err
	}
	var dummy int
	// check if sesssion exists and is not expired
	query := "SELECT 1 FROM tokens WHERE session_id = $1 AND expiry > NOW()"
	err = db.QueryRow(query, hashedSessionID).Scan(&dummy)
	if err != nil {
		if err == sql.ErrNoRows {
			// no match found
			return false, nil
		}
		// database error
		return false, err
	}
	return true, nil
}

func UserExists(db *sql.DB, email string) (bool, error) {
	var dummy int
	query := "SELECT 1 FROM users WHERE email = $1"
	err := db.QueryRow(query, email).Scan(&dummy)
	if err != nil {
		if err == sql.ErrNoRows {
			// no match found
			return false, nil
		}
		// database error
		return false, err
	}
	return true, nil
}

func UpdateAccessToken(db *sql.DB, refreshToken, sessionID string) error {
	newAccessToken, err := services.RenewAccessToken(refreshToken)
	if err != nil {
		return err
	}
	query := `UPDATE tokens SET access_token = $1, expiry = $2 WHERE session_id = $3`
	_, err = db.Exec(query, newAccessToken.AccessToken, newAccessToken.Expiry, sessionID)
	if err != nil {
		return err
	}
	return nil
}
