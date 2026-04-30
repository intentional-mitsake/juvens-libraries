package services

import (
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
)

func ProcessTokens(tokens *oauth2.Token) {
	//refreshToken := tokens.RefreshToken
	accessToken := tokens.AccessToken
	//expiry := tokens.Expiry
	GetUserInfo(accessToken)
}

func GetUserInfo(accessToken string) {
	// "https://www.googleapis.com/oauth2/v2" is the endpoint for getting user info with an access token, it has diff scopes, /userinfo is the endpoint for getting user info with an access token, it has diff scopes, we need to use the one that matches the scopes we requested in the OAuth config
	info, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil) //here we create the request to the user info endpoint, we will set the Authorization
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}
	// info is a request, we need to set the Authorization header to "Bearer <access_token>" so that the Google API knows we're authenticated and can return the user info
	// the actual user info is in the response body, we can read it and parse it as JSON to get the user's email, name, etc.
	// info --> http request from server to Google API, resp--> http response from Google API to server, we need to read the response body to get the user info
	info.Header.Set("Authorization", "Bearer "+accessToken)
	client := &http.Client{} //client is what we use to send the request to the Google API, we can use the default http client or create a new one, here we create a new one for demonstration purposes
	resp, err := client.Do(info)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("Response status: %s\n", resp.Status)
	responseData, err := io.ReadAll(resp.Body) // use io.ReadAll to read the response body, it returns a byte slice and an error, we can convert the byte slice to a string to see the user info in the response
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return
	}
	fmt.Printf("Response body: %s\n", responseData)
}
