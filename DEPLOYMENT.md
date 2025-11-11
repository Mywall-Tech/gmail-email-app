# 🚀 Email App Deployment Guide

## Quick Deployment (Recommended)

### Option 1: Railway + Vercel (Free Tier Available)

#### **Step 1: Deploy Backend to Railway**

1. **Create Railway Account**: Go to [railway.app](https://railway.app) and sign up
2. **Connect GitHub**: Link your GitHub account
3. **Push Code to GitHub**:

   ```bash
   cd /Users/vibhanshupandey/Documents/email-app
   git init
   git add .
   git commit -m "Initial commit"
   # Create a GitHub repo and push
   ```

4. **Deploy on Railway**:
   - Click "New Project" → "Deploy from GitHub repo"
   - Select your email-app repository
   - Railway will auto-detect the Dockerfile and deploy

#### **Step 2: Set Up Production Database**

1. **Add PostgreSQL**: In Railway dashboard, click "New" → "Database" → "PostgreSQL"
2. **Get Database URL**: Copy the `DATABASE_URL` from the PostgreSQL service
3. **Set Environment Variables** in Railway:

   ```
   DATABASE_URL=postgresql://postgres:password@host:port/railway
   JWT_SECRET=your_super_secure_jwt_secret_here
   GOOGLE_CLIENT_ID=74039262987-veogjg626f4d2v6cn8th1rtmv7clga1e.apps.googleusercontent.com
   GOOGLE_CLIENT_SECRET=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
   GOOGLE_REDIRECT_URL=postmessage
   GIN_MODE=release
   PORT=8080
   ```

#### **Step 3: Deploy Frontend to Vercel**

1. **Create Vercel Account**: Go to [vercel.com](https://vercel.com) and sign up
2. **Deploy Frontend**:
   - Click "New Project" → Import from GitHub
   - Select your repository
   - Set **Root Directory**: `frontend`
   - Set **Build Command**: `npm run build`
   - Set **Output Directory**: `build`
3. **Set Environment Variables** in Vercel:

   ```
   REACT_APP_API_URL=https://your-railway-app.railway.app/api
   REACT_APP_GOOGLE_CLIENT_ID=74039262987-veogjg626f4d2v6cn8th1rtmv7clga1e.apps.googleusercontent.com
   ```

#### **Step 4: Update CORS Settings**

1. **Get Vercel URL**: Copy your Vercel deployment URL
2. **Update Railway Environment**: Add `FRONTEND_URL=https://your-app.vercel.app`
3. **Redeploy**: Railway will automatically redeploy with new CORS settings

#### **Step 5: Update Google OAuth Settings**

1. **Go to Google Cloud Console**: [console.cloud.google.com](https://console.cloud.google.com)
2. **Update Authorized Origins**:
   - Add: `https://your-app.vercel.app`
   - Add: `https://your-railway-app.railway.app`
3. **Keep existing**: `postmessage` redirect URI

---

## Alternative Deployment Options

### Option 2: Docker Deployment

```bash
# Build and run with Docker
cd backend
docker build -t email-app-backend .
docker run -p 8080:8080 --env-file .env email-app-backend
```

### Option 3: Traditional VPS

1. **Set up Ubuntu/CentOS server**
2. **Install Go, Node.js, PostgreSQL**
3. **Use PM2 or systemd for process management**
4. **Set up Nginx as reverse proxy**
5. **Configure SSL with Let's Encrypt**

---

## Environment Variables Reference

### Backend (Railway)

```env
DATABASE_URL=postgresql://user:pass@host:port/db
JWT_SECRET=your-jwt-secret
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_REDIRECT_URL=postmessage
FRONTEND_URL=https://your-frontend.vercel.app
GIN_MODE=release
PORT=8080
```

### Frontend (Vercel)

```env
REACT_APP_API_URL=https://your-backend.railway.app/api
REACT_APP_GOOGLE_CLIENT_ID=your-google-client-id
```

---

## Post-Deployment Checklist

- [ ] Backend health check works: `https://your-backend.railway.app/health`
- [ ] Frontend loads: `https://your-frontend.vercel.app`
- [ ] User registration works
- [ ] User login works
- [ ] Gmail OAuth connection works
- [ ] Email sending works
- [ ] CORS is properly configured
- [ ] Database migrations ran successfully

---

## Troubleshooting

### Common Issues

1. **CORS Errors**: Make sure `FRONTEND_URL` is set in backend environment
2. **Database Connection**: Verify `DATABASE_URL` format and credentials
3. **Google OAuth**: Update redirect URIs in Google Cloud Console
4. **Build Failures**: Check that all dependencies are in `package.json`

### Logs

- **Railway**: View logs in Railway dashboard
- **Vercel**: View function logs in Vercel dashboard

---

## Cost Estimation

### Free Tier (Railway + Vercel)

- **Railway**: $0/month (500 hours free)
- **Vercel**: $0/month (100GB bandwidth)
- **Total**: $0/month for small usage

### Paid Tier

- **Railway**: $5/month (unlimited hours)
- **Vercel Pro**: $20/month (1TB bandwidth)
- **Custom Domain**: $10-15/year
- **Total**: ~$25-30/month
