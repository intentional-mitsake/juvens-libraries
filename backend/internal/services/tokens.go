package services

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"io"
	"juvens-library/internal/auth"
	"juvens-library/internal/database"
	"log/slog"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
)

// fot the google response data
type Response struct {
	id             string
	Email          string
	Verified_email bool
	Name           string
	Given_name     string
	Family_name    string
	Picture        string
}

type UserInfo struct {
	Email           string
	Name            string
	EncAccessToken  string
	EncRefreshToken string
	Expiry          time.Time
	SessionID       string
}

func ProcessTokens(tokens *oauth2.Token) (UserInfo, error) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	salt := os.Getenv("SALT")
	refreshToken := tokens.RefreshToken
	accessToken := tokens.AccessToken
	//expiry := tokens.Expiry

	//USERINFO
	responsedata, err := GetUserInfo(accessToken)
	if err != nil {
		logger.Error("Failed to get user info", "error", err)
		return UserInfo{}, err
	}
	userinfor := Response{}
	err = json.Unmarshal(responsedata, &userinfor) // unmarshal the response data into a struct, initially it was a byte slice
	if err != nil {
		logger.Error("Failed to unmarshal user info", "error", err)
		return UserInfo{}, err
	}
	logger.Info("User info retrieved", "email", userinfor.Email, "name", userinfor.Name)

	//ENC/DEC
	encAccessToken, err := EncyptToken(accessToken, []byte(salt))
	if err != nil {
		logger.Error("Failed to encrypt access token", "error", err)
		return UserInfo{}, err
	}
	logger.Info("Access token encrypted", "token", encAccessToken)
	encRefreshToken, err := EncyptToken(refreshToken, []byte(salt))
	if err != nil {
		logger.Error("Failed to encrypt refresh token", "error", err)
		return UserInfo{}, err
	}
	logger.Info("Refresh token encrypted", "token", encRefreshToken)

	// session id
	sessionID, err := generateSecureID()
	if err != nil {
		logger.Error("Failed to generate session ID", "error", err)
		return UserInfo{}, err
	}
	return UserInfo{
		Email:           userinfor.Email,
		Name:            userinfor.Name,
		EncAccessToken:  encAccessToken,
		EncRefreshToken: encRefreshToken,
		Expiry:          tokens.Expiry,
		SessionID:       sessionID,
	}, nil
}

func GetUserInfo(accessToken string) ([]byte, error) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	// "https://www.googleapis.com/oauth2/v2" is the endpoint for getting user info with an access token, it has diff scopes, /userinfo is the endpoint for getting user info with an access token, it has diff scopes, we need to use the one that matches the scopes we requested in the OAuth config
	info, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil) //here we create the request to the user info endpoint, we will set the Authorization
	if err != nil {
		return nil, err
	}
	// info is a request, we need to set the Authorization header to "Bearer <access_token>" so that the Google API knows we're authenticated and can return the user info
	// the actual user info is in the response body, we can read it and parse it as JSON to get the user's email, name, etc.
	// info --> http request from server to Google API, resp--> http response from Google API to server, we need to read the response body to get the user info
	info.Header.Set("Authorization", "Bearer "+accessToken)
	client := &http.Client{} //client is what we use to send the request to the Google API, we can use the default http client or create a new one, here we create a new one for demonstration purposes
	resp, err := client.Do(info)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	logger.Info("Response received", "status", resp.Status)
	responseData, err := io.ReadAll(resp.Body) // use io.ReadAll to read the response body, it returns a byte slice and an error, we can convert the byte slice to a string to see the user info in the response
	if err != nil {
		return nil, err
	}
	//logger.Info("Response body", "data", string(responseData))
	return responseData, nil
}
func generateSecureID() (string, error) {
	b := make([]byte, 32)  // generate 32 random bytes, which will give us a 256-bit ID, which is sufficiently secure for session IDs
	_, err := rand.Read(b) // read random bytes into the byte slice, it returns the number of bytes read and an error, we can ignore the number of bytes read since we know it will be 32, but we need to check for errors
	if err != nil {
		return "", err
	}
	// base64 as it encodes binary data into a string format that is safe for URLs and cookies, and it also makes the session ID shorter than if we were to encode it as hex, which would be 64 characters long for 32 bytes
	return base64.URLEncoding.EncodeToString(b), nil
}

func RenewAccessToken(db *sql.DB, refreshToken, sessionID string) error {
	// refresh token is long lived, use that to get new access token and expiry
	// when logging in, we check session id and expiry, so if either no session id or expired refresh token, no login
	// this is for persistent login, we check session id and expiry on each request, if expired, we use the refresh token to get a new access token and expiry, and update the database with the new access token and expiry, if the refresh token is also expired, then we log the user out by deleting the session id from the database and clearing the cookie
	// refresh token lives untill the user revokes access or changes their password
	oauthConfig := auth.GoogleOAuth() // oauth config
	token := &oauth2.Token{
		// define a token struct with only the refrsh token
		RefreshToken: refreshToken,
	}
	// theis returns a token source, which is an interface that has a Token() method that returns a new access token and expiry time, it automatically uses the refresh token to get a new access
	response := oauthConfig.Config.TokenSource(context.Background(), token)
	newAccessToken, err := response.Token()
	if err != nil {
		return err
	}
	err = database.UpdateAccessToken(db, newAccessToken.AccessToken, sessionID, newAccessToken.Expiry)
	if err != nil {
		return err
	}
	return nil
}
