# Email App with Gmail API Integration

A full-stack application for sending emails using Gmail API with user authentication and token management.

## Architecture

- **Backend**: Go with Gin framework
- **Frontend**: React with TypeScript
- **Database**: PostgreSQL with GORM
- **Authentication**: JWT tokens
- **Email Service**: Gmail API with OAuth2

## Features

- User registration and login
- JWT-based authentication
- Gmail OAuth integration
- Secure token storage
- Email sending functionality

## Project Structure

```
email-app/
├── backend/          # Go backend
├── frontend/         # React frontend
└── README.md
```

## Setup Instructions

### Backend Setup

1. Navigate to `backend/` directory
2. Run `go mod tidy` to install dependencies
3. Set up environment variables
4. Run `go run main.go`

### Frontend Setup

1. Navigate to `frontend/` directory
2. Run `npm install` to install dependencies
3. Run `npm start` to start development server

## Quick Start

1. **Setup Database**: Create a PostgreSQL database named `email_app_db`
2. **Configure Google OAuth**: Set up Gmail API credentials in Google Cloud Console
3. **Backend Setup**:

   ```bash
   cd backend
   cp env.example .env
   # Edit .env with your database and Google OAuth credentials
   go run main.go
   ```

4. **Frontend Setup**:

   ```bash
   cd frontend
   cp env.example .env
   # Edit .env if needed (default should work)
   npm install
   npm start
   ```

5. **Access the App**: Open `http://localhost:3000` in your browser

## Environment Variables

Create `.env` files in both backend and frontend directories with the required configuration. See `SETUP.md` for detailed instructions.
