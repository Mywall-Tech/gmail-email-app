#!/bin/bash

# Load environment variables from .env file
if [ -f .env ]; then
    echo "Loading environment variables from .env file..."
    export $(cat .env | grep -v '^#' | xargs)
    echo "Environment variables loaded successfully!"
    echo "GOOGLE_CLIENT_ID: $GOOGLE_CLIENT_ID"
    echo "GOOGLE_REDIRECT_URL: $GOOGLE_REDIRECT_URL"
else
    echo "No .env file found!"
fi

# Run the Go server
echo "Starting Go server..."
go run main.go
