package auth

import (
	"fmt"
	"juvens-library/internal/config"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func GoogleOAuth() config.App {
	err := godotenv.Load() // Load environment variables from .env file
	if err != nil {
		fmt.Println("Error loading .env file")
	}
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
	app := config.App{Config: oauth2cfg}
	return app
}
