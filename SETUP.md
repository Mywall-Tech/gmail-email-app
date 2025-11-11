# Email App Setup Instructions

## Prerequisites

1. **Go** (version 1.21 or higher)
2. **Node.js** (version 16 or higher)
3. **PostgreSQL** database
4. **Google Cloud Console** account for Gmail API

## Google Cloud Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the Gmail API:
   - Go to "APIs & Services" > "Library"
   - Search for "Gmail API" and enable it
4. Create OAuth 2.0 credentials:
   - Go to "APIs & Services" > "Credentials"
   - Click "Create Credentials" > "OAuth 2.0 Client IDs"
   - Choose "Web application"
   - Add authorized redirect URIs:
     - `http://localhost:8080/auth/google/callback`
   - Save the Client ID and Client Secret

## Database Setup

1. Install PostgreSQL if not already installed
2. Create a new database:

   ```sql
   CREATE DATABASE email_app_db;
   CREATE USER your_db_user WITH PASSWORD 'your_db_password';
   GRANT ALL PRIVILEGES ON DATABASE email_app_db TO your_db_user;
   ```

## Backend Setup

1. Navigate to the backend directory:

   ```bash
   cd backend
   ```

2. Copy the environment file and configure it:

   ```bash
   cp env.example .env
   ```

3. Edit `.env` file with your configuration:

   ```env
   # Database Configuration
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=your_db_user
   DB_PASSWORD=your_db_password
   DB_NAME=email_app_db

   # JWT Configuration
   JWT_SECRET=your_super_secret_jwt_key_here

   # Gmail API Configuration
   GOOGLE_CLIENT_ID=your_google_client_id
   GOOGLE_CLIENT_SECRET=your_google_client_secret
   GOOGLE_REDIRECT_URL=http://localhost:3000/dashboard

   # Server Configuration
   PORT=8080
   GIN_MODE=debug
   ```

4. Install dependencies and run:

   ```bash
   go mod tidy
   go run main.go
   ```

The backend server will start on `http://localhost:8080`

## Frontend Setup

1. Navigate to the frontend directory:

   ```bash
   cd frontend
   ```

2. Copy the environment file and configure it:

   ```bash
   cp env.example .env
   ```

3. Edit `.env` file:

   ```env
   REACT_APP_API_URL=http://localhost:8080/api
   ```

4. Install dependencies and run:

   ```bash
   npm install
   npm start
   ```

The frontend will start on `http://localhost:3000`

## Usage

1. **Register**: Create a new account at `http://localhost:3000/register`
2. **Login**: Sign in at `http://localhost:3000/login`
3. **Connect Gmail**: In the dashboard, click "Connect Gmail Account" to authorize the app
4. **Send Emails**: Once connected, use the email form to send emails through your Gmail account

## API Endpoints

### Authentication

- `POST /api/auth/register` - Register a new user
- `POST /api/auth/login` - Login user
- `GET /api/profile` - Get user profile (protected)

### Gmail Integration

- `GET /api/gmail/auth-url` - Get Gmail OAuth URL (protected)
- `GET /api/gmail/status` - Check Gmail connection status (protected)
- `POST /api/gmail/send` - Send email (protected)
- `GET /auth/google/callback` - Gmail OAuth callback

## Security Notes

1. **JWT Secret**: Use a strong, random JWT secret in production
2. **Database**: Use strong database credentials
3. **HTTPS**: Use HTTPS in production
4. **Environment Variables**: Never commit `.env` files to version control
5. **CORS**: Configure CORS properly for production domains

## Troubleshooting

1. **Database Connection Issues**: Ensure PostgreSQL is running and credentials are correct
2. **Gmail OAuth Issues**: Verify redirect URIs in Google Cloud Console
3. **CORS Issues**: Check that the frontend URL is allowed in CORS configuration
4. **Token Expiration**: Gmail tokens expire; users need to reconnect periodically
