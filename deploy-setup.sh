#!/bin/bash

echo "🚀 Email App Deployment Setup"
echo "=============================="

# Check if git is initialized
if [ ! -d ".git" ]; then
    echo "📦 Initializing Git repository..."
    git init
    
    # Create .gitignore
    cat > .gitignore << EOF
# Environment files
.env
.env.local
.env.production.local

# Dependencies
node_modules/
backend/email-app-server

# Build outputs
frontend/build/
backend/main

# Logs
*.log
npm-debug.log*

# IDE
.vscode/
.idea/

# OS
.DS_Store
Thumbs.db
EOF

    echo "✅ Git repository initialized"
else
    echo "✅ Git repository already exists"
fi

# Add all files
echo "📁 Adding files to git..."
git add .

# Commit
echo "💾 Creating initial commit..."
git commit -m "Initial commit - Email app with Gmail integration

Features:
- User authentication (register/login)
- Gmail OAuth integration
- Email sending via Gmail API
- PostgreSQL database
- React frontend with TypeScript
- Go backend with Gin framework
- JWT authentication
- Persistent login
- Production-ready deployment config"

echo ""
echo "🎉 Setup complete!"
echo ""
echo "Next steps:"
echo "1. Create a GitHub repository"
echo "2. Add remote: git remote add origin https://github.com/yourusername/email-app.git"
echo "3. Push code: git push -u origin main"
echo "4. Deploy backend to Railway: https://railway.app"
echo "5. Deploy frontend to Vercel: https://vercel.com"
echo ""
echo "📖 See DEPLOYMENT.md for detailed instructions"
