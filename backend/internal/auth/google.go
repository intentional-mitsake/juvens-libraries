package auth

import (
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func GoogleOAuth() string {
	clientID := os.Getenv("G_CLIENT_ID")
	clientSecret := os.Getenv("G_CLIENT_SECRET")
	redirectURL := os.Getenv("G_REDIRECT_URL")
	oauth2cfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  redirectURL,
		Scopes:       []string{"email", "profile"},
	}
	url := oauth2cfg.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	return url
}
