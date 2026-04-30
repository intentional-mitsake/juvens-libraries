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
	services.ProcessTokens(token)
}
