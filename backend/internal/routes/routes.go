package routes

import (
	"context"
	"fmt"
	"juvens-library/internal/auth"
	"juvens-library/internal/services"
	"net/http"

	"golang.org/x/oauth2"
)

func CreateRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/auth/oauth", oauthHandler)
	mux.HandleFunc("/auth/callback", callbackHandler)
	return mux
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprintln(w, "Welcome")
	tmpl := "../public/index.html" //use diff url for prod, this is for dev
	http.ServeFile(w, r, tmpl)
}

func oauthHandler(w http.ResponseWriter, r *http.Request) {
	app := auth.GoogleOAuth()
	// offline access type means we want a refresh token, which allows us to get a new access token when the old one expires without user interaction
	url := app.Config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	//fmt.Fprintln(w, url)
	// redirect the user to the Google OAuth URL
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	app := auth.GoogleOAuth()
	// get the code from the query parameters
	code := r.URL.Query().Get("code")
	token, err := app.Config.Exchange(context.Background(), code)
	if err != nil {
		fmt.Fprintf(w, "Failed to exchange token: %v", err)
		return
	}
	// token is a struct that contains the access token, refresh token, expiry time, etc.
	//fmt.Fprintf(w, "Token: %v", token)
	//encrypt the access sdn refresh tokens, save them in the database, and set a cookie with the user ID or session ID
	// redirect the user to the home page and set a cookie with the user ID or session ID
	sessionID, err := services.ProcessTokens(token)
	if err != nil {
		fmt.Fprintf(w, "Failed to process tokens: %v", err)
		return
	}
	fmt.Printf("Session ID: %s", sessionID)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID, // raw session id on cookie, hashed session id in database
		HttpOnly: true,      // this means the cookie cannot be accessed by JavaScript, which helps prevent XSS attacks from stealing the session ID
		Secure:   true,      // this means the cookie will only be sent over HTTPS, which helps prevent man-in-the-middle attacks from stealing the session ID, make sure to use HTTPS in production
	}) // basically we create a cookie attached to the w (the response to browser)
	hashedSessionID, err := services.HashSessionID(sessionID)
	if err != nil {
		fmt.Fprintf(w, "Failed to hash session ID: %v", err)
		return
	}
	fmt.Printf("Hashed Session ID: %s", hashedSessionID)
	// redirect the user to the home page
	http.Redirect(w, r, "/", http.StatusPermanentRedirect) // redirect user to "/" after setting the cookie, the browser will include the cookie in the request to "/", so we can use it to identify the user and show them their personalized content
}
