# Email App Backend - Gmail API Integration

Backend API service for sending emails using Gmail API with user authentication and token management.

## Architecture

- **Backend**: Go with Gin framework
- **Database**: PostgreSQL with GORM
- **Authentication**: JWT tokens
- **Email Service**: Gmail API with OAuth2

## Features

- User registration and login API endpoints
- JWT-based authentication middleware
- Gmail OAuth integration
- Secure token storage and management
- Email sending functionality via Gmail API
- Bulk email processing
- RESTful API design

## Project Structure

```
email-app-backend/
├── config/           # Database configuration
├── handlers/         # HTTP request handlers
├── middleware/       # Authentication middleware
├── models/          # Database models
├── routes/          # API route definitions
├── utils/           # Utility functions (JWT, OAuth, etc.)
├── main.go          # Application entry point
├── go.mod           # Go module dependencies
├── database-setup.sql
├── deploy-setup.sh
├── DEPLOYMENT.md
├── SETUP.md
└── README.md
```

## Setup Instructions

### Prerequisites

- Go 1.19 or higher
- PostgreSQL database
- Google Cloud Console project with Gmail API enabled

### Backend Setup

1. Clone the repository
2. Run `go mod tidy` to install dependencies
3. Set up environment variables (copy `env.example` to `.env`)
4. Run `go run main.go`

## Quick Start

1. **Setup Database**: Create a PostgreSQL database named `email_app_db`
2. **Configure Google OAuth**: Set up Gmail API credentials in Google Cloud Console
3. **Backend Setup**:

   ```bash
   git clone <repository-url>
   cd email-app-backend
   cp env.example .env
   # Edit .env with your database and Google OAuth credentials
   go mod tidy
   go run main.go
   ```

4. **API Access**: The backend API will be available at `http://localhost:8080`

## API Endpoints

- `POST /api/auth/register` - User registration
- `POST /api/auth/login` - User login
- `GET /api/auth/google` - Google OAuth initiation
- `GET /api/auth/google/callback` - Google OAuth callback
- `POST /api/gmail/send` - Send email via Gmail API
- `POST /api/gmail/bulk-send` - Send bulk emails

## Environment Variables

Create a `.env` file in the root directory with the required configuration. See `SETUP.md` for detailed instructions.

## Frontend

The frontend application has been moved to a separate repository: [email-gmail-app-fe](https://github.com/Mywall-Tech/email-gmail-app-fe)
