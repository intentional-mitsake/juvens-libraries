package routes

import (
	"database/sql"
	"juvens-library/internal/handlers"
	"net/http"
)

func CreateRouter(dObj *sql.DB) http.Handler {
	r := &handlers.Router{DB: dObj}
	mux := http.NewServeMux()
	// INDEX
	mux.HandleFunc("/", r.IndexHandler)
	// AUTH
	mux.HandleFunc("/auth", r.LoginHandler)
	mux.HandleFunc("/auth/oauth", r.OauthHandler)
	mux.HandleFunc("/auth/callback", r.CallbackHandler)
	mux.HandleFunc("/auth/logout", r.LogoutHandler)
	// LIBRARY
	mux.HandleFunc("GET /library", r.LibHandler)
	mux.HandleFunc("POST /library/{book_id}/start", r.StartHandler)
	mux.HandleFunc("POST /library/{book_id}/finish", r.FinishHandler)
	mux.HandleFunc("GET /books/search", r.SearchHandler)
	// GROUP
	mux.HandleFunc("POST /groups/{book_id}/match", r.MatchHandler)
	mux.HandleFunc("POST /groups/{group_id}/shuffle", r.ShuffleHandler)
	mux.HandleFunc("GET /groups/{group_id}/members", r.MembersHandler)
	mux.HandleFunc("GET /groups/{group_id}/messages", r.MessagesHandler)
	mux.HandleFunc("DELETE /groups/{group_id}/leave", r.LeaveHandler)
	mux.HandleFunc("PATCH /groups/{group_id}/stay", r.StayHandler)
	//

	return mux
}
