package handlers

import (
	"context"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"email-app-backend/config"
	"email-app-backend/models"
	"email-app-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

	// Track email history
	emailHistory := models.EmailHistory{
		UserID:         userID.(uint),
		EmailType:      "single",
		RecipientEmail: req.To,
		RecipientName:  "", // Single emails don't have names
		Subject:        req.Subject,
		Body:           req.Body,
		Status:         "sent",
		ErrorMessage:   "",
		BatchID:        "",
		SentAt:         time.Now(),
	}

	if err != nil {
		emailHistory.Status = "failed"
		emailHistory.ErrorMessage = err.Error()

		// Save failed email to history
		config.DB.Create(&emailHistory)

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email"})
		return
	}

	// Save successful email to history
	config.DB.Create(&emailHistory)

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

	// Always return expired as false for Gmail tokens
	c.JSON(http.StatusOK, gin.H{
		"connected":  true,
		"expires_at": gmailToken.ExpiresAt,
		"expired":    false,
		"scope":      gmailToken.Scope,
	})
}

func DisconnectGmail(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Delete all Gmail tokens for this user
	result := config.DB.Where("user_id = ?", userID).Delete(&models.GmailToken{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disconnect Gmail account"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "No Gmail account was connected",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Gmail account disconnected successfully",
		"connected": false,
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

// BulkEmailRecord represents a single email record
type BulkEmailRecord struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// ProcessCSVResponse represents the response after processing CSV
type ProcessCSVResponse struct {
	TotalRecords int               `json:"total_records"`
	ValidEmails  []BulkEmailRecord `json:"valid_emails"`
	Errors       []string          `json:"errors,omitempty"`
}

// BulkEmailRequest represents the request for bulk email sending
type BulkEmailRequest struct {
	Subject string            `json:"subject" binding:"required"`
	Body    string            `json:"body" binding:"required"`
	Emails  []BulkEmailRecord `json:"emails"`
}

// BulkEmailResponse represents the response for bulk email sending
type BulkEmailResponse struct {
	TotalEmails    int               `json:"total_emails"`
	SuccessCount   int               `json:"success_count"`
	FailureCount   int               `json:"failure_count"`
	Results        []BulkEmailResult `json:"results"`
	ProcessingTime string            `json:"processing_time"`
}

// BulkEmailResult represents the result of sending a single email
type BulkEmailResult struct {
	Email   string `json:"email"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// isValidEmail validates email format using regex
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// ProcessCSV handles CSV file upload and processing
func ProcessCSV(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Parse multipart form
	file, header, err := c.Request.FormFile("csv_file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No CSV file provided"})
		return
	}
	defer file.Close()

	// Validate file type
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".csv") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File must be a CSV file"})
		return
	}

	// Limit file size to 5MB
	const maxFileSize = 5 * 1024 * 1024
	if header.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size must be less than 5MB"})
		return
	}

	// Read and parse CSV
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	var validEmails []BulkEmailRecord
	var errors []string
	totalRecords := 0

	// Read header row
	headers, err := reader.Read()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read CSV headers"})
		return
	}

	// Find column indices
	emailCol, nameCol := -1, -1
	for i, header := range headers {
		switch strings.ToLower(strings.TrimSpace(header)) {
		case "email", "email_address", "to":
			emailCol = i
		case "name", "full_name", "recipient_name":
			nameCol = i
		}
	}

	if emailCol == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV must contain an 'email' column"})
		return
	}

	// Process data rows
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			errors = append(errors, fmt.Sprintf("Error reading row %d: %v", totalRecords+2, err))
			continue
		}

		totalRecords++

		// Skip empty rows
		if len(record) == 0 || (len(record) == 1 && strings.TrimSpace(record[0]) == "") {
			continue
		}

		// Extract email
		if emailCol >= len(record) || strings.TrimSpace(record[emailCol]) == "" {
			errors = append(errors, fmt.Sprintf("Row %d: Missing email address", totalRecords+1))
			continue
		}

		email := strings.TrimSpace(record[emailCol])
		if !isValidEmail(email) {
			errors = append(errors, fmt.Sprintf("Row %d: Invalid email format: %s", totalRecords+1, email))
			continue
		}

		// Extract name field
		name := ""
		if nameCol != -1 && nameCol < len(record) {
			name = strings.TrimSpace(record[nameCol])
		}

		validEmails = append(validEmails, BulkEmailRecord{
			Email: email,
			Name:  name,
		})
	}

	// Limit number of emails to prevent abuse
	const maxEmails = 100
	if len(validEmails) > maxEmails {
		validEmails = validEmails[:maxEmails]
		errors = append(errors, fmt.Sprintf("Limited to first %d emails", maxEmails))
	}

	fmt.Printf("User %v processed CSV: %d total records, %d valid emails, %d errors\n",
		userID, totalRecords, len(validEmails), len(errors))

	c.JSON(http.StatusOK, ProcessCSVResponse{
		TotalRecords: totalRecords,
		ValidEmails:  validEmails,
		Errors:       errors,
	})
}

// SendBulkEmails handles sending multiple emails
func SendBulkEmails(c *gin.Context) {
	startTime := time.Now()

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	var req BulkEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Emails) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No emails provided"})
		return
	}

	// Limit number of emails
	const maxBulkEmails = 100
	if len(req.Emails) > maxBulkEmails {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Maximum %d emails allowed per batch", maxBulkEmails)})
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

	// Get user email for "from" field (not currently used but available for future features)
	_, _ = c.Get("user_email")

	// Generate batch ID for grouping bulk emails
	batchID := uuid.New().String()

	// Process emails concurrently with rate limiting
	const maxConcurrent = 5
	semaphore := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	var mu sync.Mutex

	results := make([]BulkEmailResult, len(req.Emails))
	successCount := 0
	failureCount := 0

	for i, emailRecord := range req.Emails {
		wg.Add(1)
		go func(index int, record BulkEmailRecord) {
			defer wg.Done()

			// Rate limiting
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Add delay between emails to respect Gmail limits
			if index > 0 {
				time.Sleep(100 * time.Millisecond)
			}

			success := true
			errorMsg := ""

			// Personalize email body and subject if name is provided
			personalizedBody := req.Body
			personalizedSubject := req.Subject
			if record.Name != "" {
				personalizedBody = strings.ReplaceAll(personalizedBody, "{{name}}", record.Name)
				personalizedBody = strings.ReplaceAll(personalizedBody, "{{Name}}", record.Name)
				personalizedSubject = strings.ReplaceAll(personalizedSubject, "{{name}}", record.Name)
				personalizedSubject = strings.ReplaceAll(personalizedSubject, "{{Name}}", record.Name)
			}

			// Validate email
			if !isValidEmail(record.Email) {
				success = false
				errorMsg = "Invalid email format"
			} else {
				// Create email message
				emailBody := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s",
					record.Email, personalizedSubject, personalizedBody)
				message := &gmail.Message{
					Raw: base64.URLEncoding.EncodeToString([]byte(emailBody)),
				}

				// Send email
				_, err := gmailService.Users.Messages.Send("me", message).Do()
				if err != nil {
					success = false
					errorMsg = fmt.Sprintf("Failed to send: %v", err)
				}
			}

			// Track email history
			emailHistory := models.EmailHistory{
				UserID:         userID.(uint),
				EmailType:      "bulk",
				RecipientEmail: record.Email,
				RecipientName:  record.Name,
				Subject:        req.Subject,
				Body:           personalizedBody,
				Status:         "sent",
				ErrorMessage:   "",
				BatchID:        batchID,
				SentAt:         time.Now(),
			}

			if !success {
				emailHistory.Status = "failed"
				emailHistory.ErrorMessage = errorMsg
			}

			// Save to database (non-blocking)
			go func(history models.EmailHistory) {
				config.DB.Create(&history)
			}(emailHistory)

			// Update results
			mu.Lock()
			results[index] = BulkEmailResult{
				Email:   record.Email,
				Success: success,
				Error:   errorMsg,
			}
			if success {
				successCount++
			} else {
				failureCount++
			}
			mu.Unlock()

		}(i, emailRecord)
	}

	wg.Wait()

	processingTime := time.Since(startTime)

	fmt.Printf("User %v sent bulk emails: %d total, %d success, %d failed, took %v\n",
		userID, len(req.Emails), successCount, failureCount, processingTime)

	c.JSON(http.StatusOK, BulkEmailResponse{
		TotalEmails:    len(req.Emails),
		SuccessCount:   successCount,
		FailureCount:   failureCount,
		Results:        results,
		ProcessingTime: processingTime.String(),
	})
}

// EmailHistoryResponse represents paginated email history
type EmailHistoryResponse struct {
	History    []models.EmailHistory `json:"history"`
	TotalCount int64                 `json:"total_count"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	TotalPages int                   `json:"total_pages"`
}

// EmailHistoryStats represents email statistics
type EmailHistoryStats struct {
	TotalSent       int64 `json:"total_sent"`
	TotalFailed     int64 `json:"total_failed"`
	SingleEmails    int64 `json:"single_emails"`
	BulkEmails      int64 `json:"bulk_emails"`
	Last7DaysSent   int64 `json:"last_7_days_sent"`
	Last7DaysFailed int64 `json:"last_7_days_failed"`
}

// GetEmailHistory retrieves paginated email history for a user
func GetEmailHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Parse query parameters
	page := 1
	pageSize := 20
	emailType := c.Query("type") // "single", "bulk", or empty for all

	if p := c.Query("page"); p != "" {
		if parsed, err := fmt.Sscanf(p, "%d", &page); err != nil || parsed != 1 || page < 1 {
			page = 1
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := fmt.Sscanf(ps, "%d", &pageSize); err != nil || parsed != 1 || pageSize < 1 || pageSize > 100 {
			pageSize = 20
		}
	}

	// Build query
	query := config.DB.Where("user_id = ?", userID)
	if emailType != "" {
		query = query.Where("email_type = ?", emailType)
	}

	// Get total count
	var totalCount int64
	query.Model(&models.EmailHistory{}).Count(&totalCount)

	// Get paginated results
	var history []models.EmailHistory
	offset := (page - 1) * pageSize
	query.Order("sent_at DESC").Limit(pageSize).Offset(offset).Find(&history)

	totalPages := int((totalCount + int64(pageSize) - 1) / int64(pageSize))

	c.JSON(http.StatusOK, EmailHistoryResponse{
		History:    history,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// GetEmailHistoryStats retrieves email statistics for a user
func GetEmailHistoryStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	var stats EmailHistoryStats

	// Total sent and failed
	config.DB.Model(&models.EmailHistory{}).Where("user_id = ? AND status = ?", userID, "sent").Count(&stats.TotalSent)
	config.DB.Model(&models.EmailHistory{}).Where("user_id = ? AND status = ?", userID, "failed").Count(&stats.TotalFailed)

	// Single vs bulk emails
	config.DB.Model(&models.EmailHistory{}).Where("user_id = ? AND email_type = ?", userID, "single").Count(&stats.SingleEmails)
	config.DB.Model(&models.EmailHistory{}).Where("user_id = ? AND email_type = ?", userID, "bulk").Count(&stats.BulkEmails)

	// Last 7 days
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	config.DB.Model(&models.EmailHistory{}).Where("user_id = ? AND status = ? AND sent_at >= ?", userID, "sent", sevenDaysAgo).Count(&stats.Last7DaysSent)
	config.DB.Model(&models.EmailHistory{}).Where("user_id = ? AND status = ? AND sent_at >= ?", userID, "failed", sevenDaysAgo).Count(&stats.Last7DaysFailed)

	c.JSON(http.StatusOK, stats)
}
