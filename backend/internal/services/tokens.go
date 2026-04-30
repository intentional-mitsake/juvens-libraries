package services

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"

	"golang.org/x/oauth2"
)

func ProcessTokens(tokens *oauth2.Token) (string, error) {
	salt := os.Getenv("SALT")
	refreshToken := tokens.RefreshToken
	accessToken := tokens.AccessToken
	//expiry := tokens.Expiry
	_, err := GetUserInfo(accessToken)
	if err != nil {
		fmt.Printf("Error fetching user info: %v\n", err)
		return "", err
	}
	encAccessToken, err := EncyptToken(accessToken, []byte(salt))
	if err != nil {
		fmt.Printf("Error encrypting access token: %v\n", err)
		return "", err
	}
	fmt.Printf("Encrypted Access Token: %s\n", encAccessToken)
	encRefreshToken, err := EncyptToken(refreshToken, []byte(salt))
	if err != nil {
		fmt.Printf("Error encrypting refresh token: %v\n", err)
		return "", err
	}
	fmt.Printf("Encrypted Refresh Token: %s\n", encRefreshToken)
	// session id
	sessionID, err := generateSecureID()
	if err != nil {
		fmt.Printf("Error generating session ID: %v\n", err)
		return "", err
	}
	return sessionID, nil
}

func GetUserInfo(accessToken string) (string, error) {
	// "https://www.googleapis.com/oauth2/v2" is the endpoint for getting user info with an access token, it has diff scopes, /userinfo is the endpoint for getting user info with an access token, it has diff scopes, we need to use the one that matches the scopes we requested in the OAuth config
	info, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil) //here we create the request to the user info endpoint, we will set the Authorization
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return "", err
	}
	// info is a request, we need to set the Authorization header to "Bearer <access_token>" so that the Google API knows we're authenticated and can return the user info
	// the actual user info is in the response body, we can read it and parse it as JSON to get the user's email, name, etc.
	// info --> http request from server to Google API, resp--> http response from Google API to server, we need to read the response body to get the user info
	info.Header.Set("Authorization", "Bearer "+accessToken)
	client := &http.Client{} //client is what we use to send the request to the Google API, we can use the default http client or create a new one, here we create a new one for demonstration purposes
	resp, err := client.Do(info)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return "", err
	}
	defer resp.Body.Close()
	fmt.Printf("Response status: %s\n", resp.Status)
	responseData, err := io.ReadAll(resp.Body) // use io.ReadAll to read the response body, it returns a byte slice and an error, we can convert the byte slice to a string to see the user info in the response
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return "", err
	}
	fmt.Printf("Response body: %s\n", responseData)
	return string(responseData), nil
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
