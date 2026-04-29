package routes

import (
	"fmt"
	"juvens-library/internal/auth"
	"net/http"
)

func CreateRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/auth/oauth", oauthHandler)
	mux.HandleFunc("/auth/callback", callbackHandler)
	return mux
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome")
}

func oauthHandler(w http.ResponseWriter, r *http.Request) {
	url := auth.GoogleOAuth()
	fmt.Fprintln(w, url)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "callback")
}
