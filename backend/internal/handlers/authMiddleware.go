package handlers

import (
	"juvens-library/internal/database"
	"juvens-library/internal/services"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func (rt *Router) SessionValidation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
		//fmt.Fprintln(w, "Welcome")
		//cookie check
		cookie, err := r.Cookie("session_id")
		// IF ERROR OR NO COOKIE, REDIRECT TO LOGIN PAGE
		if err != nil || cookie.Value == "" {
			logger.Info("Cannot find the Session ID Cookie", "session_id", err) // if its empty, err will be nil prob. not really an error so using info level log
			// if no cookie, redirect to the login page(/auth) with a login button that redirects to /auth/oauth
			http.Redirect(w, r, "/auth", http.StatusTemporaryRedirect)
			return
		} else {
			// IF COOKIE, VALIDATE COOKIE
			logger.Info("Session ID cookie found", "session_id", cookie.Value)
			//fmt.Println("Session ID from cookie:", cookie.Value)
			hashedSessionID, err := services.HashSessionID(cookie.Value)
			//fmt.Println("Hashed Session ID:", hashedSessionID)
			if err != nil {
				logger.Error("Failed to hash session ID", "error", err)
				http.Redirect(w, r, "/auth", http.StatusTemporaryRedirect)
				return
			}
			refreshToken, expiry, exists, err := database.ValidateSessionID(rt.DB, hashedSessionID)
			if err != nil { // error in validating session ID, which likely means a database error
				logger.Error("Failed to validate session ID", "error", err)
				http.Redirect(w, r, "/auth", http.StatusTemporaryRedirect)
				return
			}
			if !exists {
				// IF NOT EXPIRED, BUT NOT EXISTS, REDIRECT TO LOGIN
				logger.Info("Session ID not found in DB", "session_id", cookie.Value)
				http.Redirect(w, r, "/auth", http.StatusTemporaryRedirect)
				return
			} else if time.Now().After(expiry) { // sessionID exists but expired
				logger.Info("Session ID expired", "session_id", cookie.Value)
				logger.Info("Attempting to renew access token using refresh token", "session_id", cookie.Value)
				salt := os.Getenv("SALT")
				//DECRYPT RTOKEN, RENEW ACCESS TOKEN, UPDATE DB, REDIRECT TO LOGIN IF ANY OF THESE STEPS FAIL, ELSE GET OUT OF IF STATEMENT
				decryptedRefreshToken, err := services.DecryptToken(refreshToken, []byte(salt))
				if err != nil {
					logger.Error("Failed to decrypt refresh token", "error", err)
					http.Redirect(w, r, "/auth", http.StatusTemporaryRedirect)
					return
				}
				err = services.RenewAccessToken(rt.DB, decryptedRefreshToken, hashedSessionID)
				if err != nil {
					logger.Error("Failed to renew access token", "error", err)
					http.Redirect(w, r, "/auth", http.StatusTemporaryRedirect)
					return
				}
				return
			}
			// REACH HERE IF SESSION ID VALID AND NOT EXPIRED, REDIRECT TO NEXT HANDLER
			logger.Info("Valid session ID, loading index page", "session_id", cookie.Value)
			next.ServeHTTP(w, r)
		}
	})
}
