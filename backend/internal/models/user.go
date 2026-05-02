package models

import "time"

type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type Token struct {
	UserID       string    `json:"user_id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
	SessionID    string    `json:"session_id"`
}
type Book struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

type UserBook struct {
	UserID string `json:"user_id"`
	BookID string `json:"book_id"`
	// use pointers for times so they can be "nil" if the book isn't started/finished
	StartedAt  *time.Time `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
	Rating     int        `json:"rating"`
	ReadStatus int        `json:"rstatus"` // 0 = not started, 1 = reading, 2 = finished
}
