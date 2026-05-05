package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"juvens-library/internal/auth"
	"juvens-library/internal/database"
	"juvens-library/internal/services"
	"log/slog"
	"net/http"
	"os"

	"golang.org/x/oauth2"
)

type Router struct {
	DB *sql.DB
}

func (rt *Router) IndexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := "../public/index.html"
	http.ServeFile(w, r, tmpl)
}

func (rt *Router) LoginHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := "../public/login.html"
	http.ServeFile(w, r, tmpl)
}

func (rt *Router) OauthHandler(w http.ResponseWriter, r *http.Request) {
	app := auth.GoogleOAuth()
	// offline access type means we want a refresh token, which allows us to get a new access token when the old one expires without user interaction
	url := app.Config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	//fmt.Fprintln(w, url)
	// redirect the user to the Google OAuth URL
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (rt *Router) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	app := auth.GoogleOAuth()
	// get the code from the query parameters
	code := r.URL.Query().Get("code")
	token, err := app.Config.Exchange(context.Background(), code)
	if err != nil {
		logger.Error("Failed to exchange code for token", "error", err)
		http.Redirect(w, r, "/auth", http.StatusTemporaryRedirect) //if exchange fails, redirect to login
		return
	}
	// token is a struct that contains the access token, refresh token, expiry time, etc.
	//fmt.Fprintf(w, "Token: %v", token)
	//encrypt the access sdn refresh tokens, save them in the database, and set a cookie with the user ID or session ID
	// redirect the user to the home page and set a cookie with the user ID or session ID
	userinfo, err := services.ProcessTokens(token)
	if err != nil {
		logger.Error("Failed to process tokens", "error", err)
		http.Redirect(w, r, "/auth", http.StatusTemporaryRedirect)
		return
	}
	logger.Info("Session ID created")
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    userinfo.SessionID, // raw session id on cookie, hashed session id in database
		Path:     "/",                // cookie is visible to /
		MaxAge:   86400,
		HttpOnly: true,  // this means the cookie cannot be accessed by JavaScript, which helps prevent XSS attacks from stealing the session ID
		Secure:   false, // this means the cookie will only be sent over HTTPS, which helps prevent man-in-the-middle attacks from stealing the session ID, make sure to use HTTPS in production
	}) // basically we create a cookie attached to the w (the response to browser)
	hashedSessionID, err := services.HashSessionID(userinfo.SessionID)
	if err != nil {
		logger.Error("Failed to hash session ID", "error", err)
		http.Redirect(w, r, "/auth", http.StatusTemporaryRedirect)
		return
	}
	olduser, err := database.UserExists(rt.DB, userinfo.Email)
	if err != nil {
		logger.Error("Failed to check if user exists", "error", err)
		http.Redirect(w, r, "/auth", http.StatusTemporaryRedirect)
		return
	}
	if !olduser {
		//i maybe a tutorial here or a welcome message for new users, but for now we will just log it
		logger.Info("New user logged in", "email", userinfo.Email)

	}
	err = database.InsertLoginInfo(rt.DB, userinfo.Email, userinfo.Name, userinfo.EncAccessToken, userinfo.EncRefreshToken, hashedSessionID, userinfo.Expiry)
	if err != nil {
		logger.Error("Failed to insert login info", "error", err)
		http.Redirect(w, r, "/auth", http.StatusTemporaryRedirect)
		return
	}
	// redirect the user to the home page
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect) // redirect user to "/" after setting the cookie, the browser will include the cookie in the request to "/", so we can use it to identify the user and show them their personalized content
}

func (rt *Router) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	cookie, err := r.Cookie("session_id")
	if err != nil {
		logger.Error("Failed to get session ID cookie", "error", err)
		fmt.Fprintln(w, "Error logging out") // writes to the resp body, appears on page
		return
	}
	//fmt.Println("Session ID from cookie during logout:", cookie.Value)
	hashedSessionID, err := services.HashSessionID(cookie.Value)
	//fmt.Println("Hashed session ID during logout:", hashedSessionID)
	if err != nil {
		logger.Error("Failed to hash session ID", "error", err)
		fmt.Fprintln(w, "Error logging out")
		return
	}
	err = database.RevokeSession(rt.DB, hashedSessionID)
	if err != nil {
		logger.Error("Failed to revoke session", "error", err)
		fmt.Fprintln(w, "Error logging out")
		return
	}
	// delete the cookie by setting its MaxAge to -1
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		MaxAge:   -1, // this tells the browser to delete the cookie
	})
	http.Redirect(w, r, "/auth", http.StatusTemporaryRedirect) // redirect to login page after logout
	logger.Info("Session revoked successfully", "session_id", cookie.Value)
}
