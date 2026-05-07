package database

import (
	"database/sql"
	"fmt"
	"time"

	"juvens-library/internal/config"
	"juvens-library/internal/models"
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
        token_expiry TIMESTAMPTZ NOT NULL,
        session_id TEXT PRIMARY KEY,
		session_expiry TIMESTAMPTZ NOT NULL,
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

func InsertLoginInfo(db *sql.DB, email, name, encAccessToken, encRefreshToken, hashedSessionID string, token_expiry, session_expiry time.Time) error {
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
		INSERT INTO tokens (user_id, access_token, refresh_token, token_expiry, session_id, session_expiry)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (session_id) DO UPDATE 
		SET access_token = EXCLUDED.access_token, refresh_token = EXCLUDED.refresh_token, expiry = EXCLUDED.expiry;
	`
	var generatedUserID string                                       // to store the new or exsitng id
	err = tx.QueryRow(userquery, email, name).Scan(&generatedUserID) // queeryrow as we need to get the generated id
	if err != nil {
		tx.Rollback() // if there is an error in the transaction, we need to rollback to maintain data integrity
		return err
	}
	_, err = tx.Exec(tokenquery, generatedUserID, encAccessToken, encRefreshToken, token_expiry, hashedSessionID, session_expiry)
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

func ValidateSessionID(db *sql.DB, sessionID string) (string, time.Time, time.Time, bool, error) {
	query := `SELECT refresh_token, session_expiry, token_expiry FROM tokens WHERE session_id = $1`
	var refreshToken string
	var session_expiry time.Time
	var token_expiry time.Time
	err := db.QueryRow(query, sessionID).Scan(&refreshToken, &session_expiry, &token_expiry)
	if err != nil {
		if err == sql.ErrNoRows {
			// no match found, session ID is not valid
			return "", time.Time{}, time.Time{}, false, nil
		}
		return "", time.Time{}, time.Time{}, false, err
	}
	fmt.Println(refreshToken)
	return refreshToken, session_expiry, token_expiry, true, nil
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

func UpdateSessionState(db *sql.DB, sessionID, accessToken string, tokenExpiry, sessionExpiry time.Time) error {
	// We update everything in one go: the key, the short life, and the long life.
	query := `
        UPDATE tokens 
        SET access_token = $1, 
            token_expiry = $2, 
            session_expiry = $3 
        WHERE session_id = $4`

	_, err := db.Exec(query, accessToken, tokenExpiry, sessionExpiry, sessionID)
	return err
}

func RevokeSession(db *sql.DB, sessionID string) error {
	//fmt.Println("Revoking session with ID:", sessionID) // Debug log to check the session ID being revoked
	query := `DELETE FROM tokens WHERE session_id = $1`
	_, err := db.Exec(query, sessionID)
	if err != nil {
		return err
	}
	/* for debugging only, there was a bug where session id from cookie was not matching the one in the database,
		this helped see that it was not effecting any rows--> so no matching
		no need to keep this here tho as if no rows are effected, that means there is no row to delete, so can ignore that
		as if there was a session to delete, it would have been deleted already
	// Check if any rows were affected
	if rowsAffected, err := res.RowsAffected(); err != nil {
		return err
	} else if rowsAffected == 0 {
		// No rows deleted, session ID not found
		return sql.ErrNoRows
	}
	*/
	return nil
}

func GetUserLibrary(db *sql.DB, userID string) ([]models.UserBook, error) {
	query := `SELECT * FROM user_books WHERE user_id = $1`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	i := 0
	userbooks := make([]models.UserBook, 0, 100) // prealloactes an array of size 100 with zero values
	for rows.Next() {
		var userbook models.UserBook
		if err := rows.Scan(
			&userbook.UserID,
			&userbook.BookID,
			&userbook.StartedAt,
			&userbook.FinishedAt,
			&userbook.Rating,
			&userbook.ReadStatus,
		); err != nil {
			return nil, err
		}
		userbooks = append(userbooks, userbook)
		i++
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return userbooks, nil
}
