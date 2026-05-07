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
	mux.HandleFunc("/", r.IndexHandler) //-->allow everyone to get here, session validation only if they touch user sensitive things
	// when they hit stuff like library, they will be validated, will use soft session check to show login/logut at index page
	// if theres session cookie, but invlaid/expired it will still show logout this way, but wont effect anything and will redirect to login anyway
	// same with lib and group stuff, so a user exp bug but saves a lot of work i dont want to do right now
	// should work better tho cuz cookie is deleted at expiry or when user logs out anyway so it should be fine
	// AUTH
	mux.HandleFunc("GET /auth", r.LoginHandler)
	mux.HandleFunc("GET /auth/oauth", r.OauthHandler)
	mux.HandleFunc("GET /auth/callback", r.CallbackHandler)
	mux.HandleFunc("POST /auth/logout", r.LogoutHandler)
	// LIBRARY
	mux.Handle("GET /library", r.SessionValidation(http.HandlerFunc(r.LibHandler)))
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
