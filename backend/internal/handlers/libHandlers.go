package handlers

import (
	"encoding/json"
	"juvens-library/internal/database"
	"juvens-library/internal/services"
	"log/slog"
	"net/http"
	"os"
)

func (rt *Router) LibHandler(w http.ResponseWriter, r *http.Request) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	user_id := r.URL.Query().Get("user_id")
	if user_id == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	library, err := database.GetUserLibrary(rt.DB, user_id)
	if err != nil {
		logger.Error("Failed to get library", "error", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(library); err != nil {
		logger.Error("Failed to encode JSON", "error", err)
	}
	logger.Info("User Library Fetched", "user_id", user_id, "library", library)
}

func (rt *Router) StartHandler(w http.ResponseWriter, r *http.Request) {

}

func (rt *Router) FinishHandler(w http.ResponseWriter, r *http.Request) {
}

func (rt *Router) BookHandler(w http.ResponseWriter, r *http.Request) {
}

func (rt *Router) SearchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("query")
	books, err := services.SearchBooks(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(books); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
