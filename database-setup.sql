-- PostgreSQL Database Setup for Email App
-- Run these commands in psql or your PostgreSQL client

-- Create the database
CREATE DATABASE email_app_db;

-- Create a user for the application (optional but recommended)
CREATE USER email_app_user WITH PASSWORD '1122@aA';

-- Grant privileges to the user
GRANT ALL PRIVILEGES ON DATABASE email_app_db TO email_app_user;

-- Connect to the database to set up additional permissions
\c email_app_db

-- Grant schema permissions (PostgreSQL 15+)
GRANT ALL ON SCHEMA public TO email_app_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO email_app_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO email_app_user;

-- Set default privileges for future tables
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO email_app_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO email_app_user;
