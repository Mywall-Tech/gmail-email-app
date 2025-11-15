package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Email     string         `json:"email" gorm:"uniqueIndex;not null"`
	Password  string         `json:"-" gorm:"not null"`
	Name      string         `json:"name" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Gmail integration
	GmailTokens  []GmailToken   `json:"gmail_tokens,omitempty" gorm:"foreignKey:UserID"`
	EmailHistory []EmailHistory `json:"email_history,omitempty" gorm:"foreignKey:UserID"`
}

type GmailToken struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	UserID       uint           `json:"user_id" gorm:"not null"`
	AccessToken  string         `json:"-" gorm:"not null"`
	RefreshToken string         `json:"-" gorm:"not null"`
	TokenType    string         `json:"token_type" gorm:"default:'Bearer'"`
	ExpiresAt    time.Time      `json:"expires_at"`
	Scope        string         `json:"scope"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationship
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type EmailHistory struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	UserID         uint           `json:"user_id" gorm:"not null"`
	EmailType      string         `json:"email_type" gorm:"not null"` // "single" or "bulk"
	RecipientEmail string         `json:"recipient_email" gorm:"not null"`
	RecipientName  string         `json:"recipient_name"`
	Subject        string         `json:"subject" gorm:"not null"`
	Body           string         `json:"body" gorm:"type:text"`
	Status         string         `json:"status" gorm:"not null"` // "sent" or "failed"
	ErrorMessage   string         `json:"error_message"`
	BatchID        string         `json:"batch_id"` // For grouping bulk emails
	SentAt         time.Time      `json:"sent_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationship
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}
