package routes

import (
	"context"
	"database/sql"
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

func CreateRouter(dObj *sql.DB) http.Handler {
	r := &Router{DB: dObj}
	mux := http.NewServeMux()
	mux.HandleFunc("/", r.indexHandler)
	mux.HandleFunc("/auth", r.loginHandler)
	mux.HandleFunc("/auth/oauth", r.oauthHandler)
	mux.HandleFunc("/auth/callback", r.callbackHandler)
	return mux
}

func (rt *Router) indexHandler(w http.ResponseWriter, r *http.Request) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	//fmt.Fprintln(w, "Welcome")
	//cookie check
	cookie, err := r.Cookie("session_id")
	if err != nil || cookie.Value == "" {
		logger.Info("Cannot find the Session ID Cookie", "session_id", err) // if its empty, err will be nil prob. not really an error so using info level log
		// if no cookie, redirect to the login page(/auth) with a login button that redirects to /auth/oauth
		http.Redirect(w, r, "/auth", http.StatusTemporaryRedirect)
	} else {
		// if there is cookie, check if its valid
		logger.Info("Session ID cookie found", "session_id", cookie)
		hashedSessionID, err := services.HashSessionID(cookie.Value)
		if err != nil {
			logger.Error("Failed to hash session ID", "error", err)
			http.Redirect(w, r, "/auth", http.StatusTemporaryRedirect)
			return
		}
		valid, err := database.ValidateSessionID(rt.DB, hashedSessionID)
		if err != nil { // error in validating session ID, which likely means a database error
			logger.Error("Failed to validate session ID", "error", err)
			http.Redirect(w, r, "/auth", http.StatusTemporaryRedirect)
			return
		}
		if !valid { // session ID is not valid, no db error, just means the session ID is not in the database or expired
			logger.Info("Invalid session ID", "session_id", cookie.Value)
			http.Redirect(w, r, "/auth", http.StatusTemporaryRedirect)
			return
		}
		// if valid, load the index page.
		logger.Info("Valid session ID, loading index page", "session_id", cookie.Value)
		tmpl := "../public/index.html" // if cookie valid, load index page
		http.ServeFile(w, r, tmpl)
	}
}

func (rt *Router) loginHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := "../public/login.html"
	http.ServeFile(w, r, tmpl)
}

func (rt *Router) oauthHandler(w http.ResponseWriter, r *http.Request) {
	app := auth.GoogleOAuth()
	// offline access type means we want a refresh token, which allows us to get a new access token when the old one expires without user interaction
	url := app.Config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	//fmt.Fprintln(w, url)
	// redirect the user to the Google OAuth URL
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (rt *Router) callbackHandler(w http.ResponseWriter, r *http.Request) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	app := auth.GoogleOAuth()
	// get the code from the query parameters
	code := r.URL.Query().Get("code")
	token, err := app.Config.Exchange(context.Background(), code)
	if err != nil {
		logger.Error("Failed to exchange code for token", "error", err)
		return
	}
	// token is a struct that contains the access token, refresh token, expiry time, etc.
	//fmt.Fprintf(w, "Token: %v", token)
	//encrypt the access sdn refresh tokens, save them in the database, and set a cookie with the user ID or session ID
	// redirect the user to the home page and set a cookie with the user ID or session ID
	userinfo, err := services.ProcessTokens(token)
	if err != nil {
		logger.Error("Failed to process tokens", "error", err)
		return
	}
	logger.Info("Session ID created")
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    userinfo.SessionID, // raw session id on cookie, hashed session id in database
		Path:     "/",                // cookie is visible to /
		HttpOnly: true,               // this means the cookie cannot be accessed by JavaScript, which helps prevent XSS attacks from stealing the session ID
		Secure:   true,               // this means the cookie will only be sent over HTTPS, which helps prevent man-in-the-middle attacks from stealing the session ID, make sure to use HTTPS in production
	}) // basically we create a cookie attached to the w (the response to browser)
	hashedSessionID, err := services.HashSessionID(userinfo.SessionID)
	if err != nil {
		logger.Error("Failed to hash session ID", "error", err)
		return
	}
	olduser, err := database.UserExists(rt.DB, userinfo.Email)
	if err != nil {
		logger.Error("Failed to check if user exists", "error", err)
		return
	}
	if !olduser {
		//i maybe a tutorial here or a welcome message for new users, but for now we will just log it
		logger.Info("New user logged in", "email", userinfo.Email)
	}
	err = database.InsertLoginInfo(rt.DB, userinfo.Email, userinfo.Name, userinfo.EncAccessToken, userinfo.EncRefreshToken, hashedSessionID, userinfo.Expiry)
	if err != nil {
		logger.Error("Failed to insert login info", "error", err)
		return
	}
	// redirect the user to the home page
	http.Redirect(w, r, "/", http.StatusPermanentRedirect) // redirect user to "/" after setting the cookie, the browser will include the cookie in the request to "/", so we can use it to identify the user and show them their personalized content
}
