package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

type GoogleTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

// GetGmailRefreshToken exchanges authorization code for access and refresh tokens
func GetGmailRefreshToken(code, redirectURL string) (*GoogleTokenResponse, error) {
	fmt.Println("GetGmailRefreshToken called with code:", code, "and redirectURL:", redirectURL)
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")

	fmt.Println("clientID:", clientID)
	fmt.Println("clientSecret:", clientSecret)

	// Validate required parameters
	if clientID == "" {
		return nil, fmt.Errorf("GOOGLE_CLIENT_ID environment variable is not set")
	}
	if clientSecret == "" {
		return nil, fmt.Errorf("GOOGLE_CLIENT_SECRET environment variable is not set")
	}
	if code == "" {
		return nil, fmt.Errorf("authorization code is empty")
	}
	if redirectURL == "" {
		return nil, fmt.Errorf("redirect URL is empty")
	}

	// Prepare form data
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("redirect_uri", redirectURL)
	data.Set("grant_type", "authorization_code")

	fmt.Printf("OAuth request parameters:\n")
	fmt.Printf("  client_id: %s\n", clientID)
	fmt.Printf("  client_secret: %s\n", clientSecret)
	fmt.Printf("  redirect_uri: %s\n", redirectURL)
	fmt.Printf("  code: %s\n", code)
	fmt.Printf("  grant_type: authorization_code\n")
	fmt.Printf("Full encoded data: %s\n", data.Encode())

	// Make the request
	resp, err := http.Post(
		"https://oauth2.googleapis.com/token",
		"application/x-www-form-urlencoded",
		bytes.NewBufferString(data.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to make token request: %v", err)
	}
	fmt.Println("resp", resp)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read the error response body for debugging
		var errorBody bytes.Buffer
		errorBody.ReadFrom(resp.Body)
		fmt.Printf("Google OAuth error response (status %d): %s\n", resp.StatusCode, errorBody.String())
		return nil, fmt.Errorf("token request failed with status: %d, body: %s", resp.StatusCode, errorBody.String())
	}

	var tokenResp GoogleTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %v", err)
	}

	return &tokenResp, nil
}
