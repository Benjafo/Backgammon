# Railway Deployment Steps

## Prerequisites
- Railway account connected to GitHub
- Repository pushed to GitHub

## Deployment Steps

### 1. Create Railway Project
```bash
railway login
railway init
```
Or via Railway dashboard: Click "New Project" → "Deploy from GitHub repo" → Select your repository

### 2. Add PostgreSQL Database
In Railway dashboard:
- Click "+ New" → "Database" → "Add PostgreSQL"
- Railway automatically creates `DATABASE_URL` variable

### 3. Initialize Database Schema
Install PostgreSQL client if not already installed:
```bash
sudo apt update
sudo apt install postgresql-client
```

After PostgreSQL is provisioned on Railway:
```bash
railway link  # Link to your Railway project (select the project when prompted)
cat schema.sql | railway run psql $DATABASE_URL
```

### 4. Configure App Service
In Railway dashboard, click on your app service:
- Go to "Settings"
- Set "Root Directory" to `/` (default)
- Set "Dockerfile Path" to `Dockerfile` (should auto-detect)
- Confirm "Port" shows `8080`

### 5. Set Environment Variables
Railway auto-injects `DATABASE_URL` from PostgreSQL service. Verify in "Variables" tab:
- `DATABASE_URL` should exist (linked from PostgreSQL)
- Modify if needed: Remove `?sslmode=disable` for production (Railway Postgres supports SSL)
- Set to: `${{Postgres.DATABASE_URL}}?sslmode=require`

### 6. Deploy
- Push to your default branch (Railway auto-deploys)
- Or click "Deploy" in Railway dashboard
- Monitor build logs in "Deployments" tab

### 7. Get Public URL
- Go to "Settings" → "Networking"
- Click "Generate Domain"
- Your app will be available at: `https://your-app.up.railway.app`

## Verify Deployment
```bash
curl https://your-app.up.railway.app/api/v1/health
```

## Troubleshooting

### Build fails at frontend stage
- Ensure `client/package.json` exists
- Check Node version in Dockerfile matches Railway's Node image

### Database connection fails
- Check `DATABASE_URL` format includes `?sslmode=require`
- Verify PostgreSQL service is running
- Ensure app service has reference to Postgres in "Variables" tab

### App crashes on startup
- Check "Deployments" logs for Go build errors
- Verify port 8080 is exposed
- Confirm schema was applied to database

## Connect to Railway Postgres Locally
```bash
railway link
railway run psql $DATABASE_URL
```
