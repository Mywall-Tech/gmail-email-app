package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"email-app-backend/config"
	"email-app-backend/models"
	"email-app-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

var (
	googleOAuthConfig *oauth2.Config
)

func init() {
	// Try to load .env file if environment variables are not set
	if os.Getenv("GOOGLE_CLIENT_ID") == "" {
		fmt.Println("Attempting to load .env file...")
		if err := godotenv.Load(); err != nil {
			fmt.Printf("Could not load .env file: %v\n", err)
		}
	}

	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURL := os.Getenv("GOOGLE_REDIRECT_URL")

	if clientID == "" {
		fmt.Printf("WARNING: GOOGLE_CLIENT_ID is not set!\n")
	}
	if clientSecret == "" {
		fmt.Printf("WARNING: GOOGLE_CLIENT_SECRET is not set!\n")
	}
	if redirectURL == "" {
		fmt.Printf("WARNING: GOOGLE_REDIRECT_URL is not set!\n")
	}

	googleOAuthConfig = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			gmail.GmailSendScope,
		},
		Endpoint: google.Endpoint,
	}

	fmt.Printf("Google OAuth Config initialized:\n")
	fmt.Printf("  Client ID: %s\n", clientID)
	secretPreview := clientSecret
	if len(clientSecret) > 10 {
		secretPreview = clientSecret[:10] + "..."
	}
	fmt.Printf("  Client Secret: %s\n", secretPreview)
	fmt.Printf("  Redirect URL: %s\n", redirectURL)
}

func GetGmailAuthURL(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Generate state parameter with user ID for security
	state := fmt.Sprintf("user_%d_%d", userID, time.Now().Unix())

	url := googleOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent"))

	c.JSON(http.StatusOK, gin.H{
		"auth_url": url,
		"state":    state,
	})
}

func HandleGmailCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code not provided"})
		return
	}

	// Exchange code for token
	token, err := googleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange code for token"})
		return
	}

	// Extract user ID from state (in production, you'd want more robust state validation)
	var userID uint
	if _, err := fmt.Sscanf(state, "user_%d_%d", &userID, new(int64)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state parameter"})
		return
	}

	// Save or update Gmail token in database
	gmailToken := models.GmailToken{
		UserID:       userID,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		ExpiresAt:    token.Expiry,
		Scope:        "gmail.send",
	}

	// Check if token already exists for this user
	var existingToken models.GmailToken
	if err := config.DB.Where("user_id = ?", userID).First(&existingToken).Error; err == nil {
		// Update existing token
		config.DB.Model(&existingToken).Updates(gmailToken)
	} else {
		// Create new token
		config.DB.Create(&gmailToken)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Gmail account connected successfully",
		"expires_at": token.Expiry,
	})
}

type SendEmailRequest struct {
	To      string `json:"to" binding:"required,email"`
	Subject string `json:"subject" binding:"required"`
	Body    string `json:"body" binding:"required"`
}

func SendEmail(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	var req SendEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user's Gmail token
	var gmailToken models.GmailToken
	if err := config.DB.Where("user_id = ?", userID).First(&gmailToken).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Gmail account not connected"})
		return
	}

	// Create OAuth2 token
	token := &oauth2.Token{
		AccessToken:  gmailToken.AccessToken,
		RefreshToken: gmailToken.RefreshToken,
		TokenType:    gmailToken.TokenType,
		Expiry:       gmailToken.ExpiresAt,
	}

	// Create Gmail service
	ctx := context.Background()
	client := googleOAuthConfig.Client(ctx, token)

	gmailService, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Gmail service"})
		return
	}

	// Get user email for "from" field
	userEmail, _ := c.Get("user_email")

	// Create email message
	emailBody := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", req.To, req.Subject, req.Body)
	message := &gmail.Message{
		Raw: base64.URLEncoding.EncodeToString([]byte(emailBody)),
	}

	// Send email
	_, err = gmailService.Users.Messages.Send("me", message).Do()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Email sent successfully",
		"to":      req.To,
		"subject": req.Subject,
		"from":    userEmail,
	})
}

func GetGmailStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	var gmailToken models.GmailToken
	if err := config.DB.Where("user_id = ?", userID).First(&gmailToken).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"connected": false,
			"message":   "Gmail account not connected",
		})
		return
	}

	// Check if token is expired
	isExpired := time.Now().After(gmailToken.ExpiresAt)

	c.JSON(http.StatusOK, gin.H{
		"connected":  true,
		"expires_at": gmailToken.ExpiresAt,
		"expired":    isExpired,
		"scope":      gmailToken.Scope,
	})
}

type GoogleCallbackRequest struct {
	Code  string `json:"code" binding:"required"`
	Scope string `json:"scope,omitempty"`
}

func HandleGoogleCallback(c *gin.Context) {
	var req GoogleCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("Received authorization code: %s\n", req.Code)
	fmt.Printf("Received scope: %s\n", req.Scope)

	// Use our custom GetGmailRefreshToken function
	redirectURL := os.Getenv("GOOGLE_REDIRECT_URL")
	if redirectURL == "" {
		redirectURL = "postmessage" // Default for LoginSocialGoogle
	}

	tokenResp, err := utils.GetGmailRefreshToken(req.Code, redirectURL)
	if err != nil {
		fmt.Printf("Token exchange error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange code for token", "details": err.Error()})
		return
	}

	fmt.Printf("Token exchange successful. Access token: %s...\n", tokenResp.AccessToken[:20])

	// Get user info from Google using the access token
	userInfoResp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + tokenResp.AccessToken)
	if err != nil {
		fmt.Printf("Failed to get user info: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}
	defer userInfoResp.Body.Close()

	var userInfo struct {
		Email string `json:"email"`
		Name  string `json:"name"`
		ID    string `json:"id"`
	}

	if err := json.NewDecoder(userInfoResp.Body).Decode(&userInfo); err != nil {
		fmt.Printf("Failed to decode user info: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode user info"})
		return
	}

	fmt.Printf("User info retrieved: %s (%s)\n", userInfo.Name, userInfo.Email)

	// Check if user exists, create if not
	var user models.User
	if err := config.DB.Where("email = ?", userInfo.Email).First(&user).Error; err != nil {
		// Create new user
		user = models.User{
			Name:     userInfo.Name,
			Email:    userInfo.Email,
			Password: "", // No password for OAuth users
		}
		if err := config.DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
		fmt.Printf("Created new user: %s\n", user.Email)
	} else {
		fmt.Printf("Found existing user: %s\n", user.Email)
	}

	// Calculate expiry time
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	// Save Gmail token to database
	gmailToken := models.GmailToken{
		UserID:       user.ID,
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    tokenResp.TokenType,
		ExpiresAt:    expiresAt,
		Scope:        req.Scope, // Use the scope from the request
	}

	// Update or create Gmail token
	var existingToken models.GmailToken
	if err := config.DB.Where("user_id = ?", user.ID).First(&existingToken).Error; err == nil {
		// Update existing token
		fmt.Printf("Updating existing Gmail token for user %d\n", user.ID)
		config.DB.Model(&existingToken).Updates(gmailToken)
	} else {
		// Create new token
		fmt.Printf("Creating new Gmail token for user %d\n", user.ID)
		config.DB.Create(&gmailToken)
	}

	// Generate JWT for the user session
	jwtToken, err := utils.GenerateJWT(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate JWT token"})
		return
	}

	fmt.Printf("Authentication successful for user: %s\n", user.Email)

	c.JSON(http.StatusOK, gin.H{
		"token":           jwtToken,
		"user":            user,
		"message":         "Gmail authentication and connection successful",
		"gmail_connected": true,
	})
}
