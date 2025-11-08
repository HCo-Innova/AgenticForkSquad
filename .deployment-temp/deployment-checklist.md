# 🚀 Deployment Checklist

## Railway Backend

### 1. Create Railway Project
- [ ] Go to https://railway.app
- [ ] Click "New Project"
- [ ] Select "Deploy from GitHub"
- [ ] Select repo: HCo-Innova/AgenticForkSquad
- [ ] Wait for build to complete

### 2. Configure Railway Variables
- [ ] Go to Project Settings → Variables
- [ ] Add all Tiger Cloud credentials (see tiger-cloud-credentials.txt)
- [ ] Add GCP_CREDENTIALS_BASE64 (see above)
- [ ] Add VERTEX_PROJECT_ID, VERTEX_LOCATION
- [ ] Set PORT=8000
- [ ] Set ENV=production
- [ ] Set LOG_LEVEL=info
- [ ] Save and redeploy

### 3. Verify Railway Deployment
- [ ] Check build status (should be green)
- [ ] Copy Railway public URL (e.g., https://afs-backend-prod.railway.app)
- [ ] Test health endpoint:
  curl https://afs-backend-prod.railway.app/health

### 4. Get Railway URL
- [ ] Note the public URL for Vercel configuration

## Vercel Frontend

### 1. Create Vercel Project
- [ ] Go to https://vercel.com
- [ ] Click "Add New" → "Project"
- [ ] Import repo: HCo-Innova/AgenticForkSquad
- [ ] Select root: ./frontend (if prompted)
- [ ] Select Build: Vite

### 2. Configure Vercel Variables
- [ ] Go to Settings → Environment Variables
- [ ] Add VITE_API_URL=https://[RAILWAY_URL]/api/v1
  (Replace [RAILWAY_URL] with actual Railway URL from step 4 above)
- [ ] Add VITE_WS_URL=wss://[RAILWAY_URL]/ws
- [ ] Add NODE_ENV=production
- [ ] Save and redeploy

### 3. Verify Vercel Deployment
- [ ] Check build status (should be green)
- [ ] Open Vercel URL in browser
- [ ] Check browser console for errors
- [ ] Verify Network tab shows API calls to Railway

## Post-Deployment Verification

### Health Checks
- [ ] Backend health: curl https://[RAILWAY_URL]/health
- [ ] Frontend loads: curl https://[VERCEL_URL]
- [ ] WebSocket connects (check browser console)
- [ ] Tiger Cloud accessible (check backend logs)
- [ ] GCP credentials working (check backend logs for Vertex AI)

### Full E2E Test
- [ ] Login to frontend
- [ ] Create optimization task
- [ ] Verify data flows through Gemini
- [ ] Check fork operations work
- [ ] Verify results displayed in UI

