package routes

import (
	"fmt"
	"net/http"
)

func CreateRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/login", loginHandler)
	return mux
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome")
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "gud")
}
