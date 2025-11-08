# 🚀 Deployment Guide: Vercel + Railway + Tiger Cloud

**Fecha**: Noviembre 2025  
**Estado**: Ready for Production  
**Arquitectura**: Frontend (Vercel) → Backend (Railway) → Database (Tiger Cloud)

---

## 📋 Tabla de Valores - Tiger Cloud Configuration

### ✅ VALORES CONFIRMADOS EN `.env`

| Variable | Valor | Destino |
|----------|-------|---------|
| **USE_TIGER_CLOUD** | `true` | ✅ Enable Tiger Cloud |
| **TIGER_PROJECT_ID** | `a1lqw18o6u` | Tiger Cloud project identifier |
| **TIGER_MAIN_SERVICE** | `wuj5xa6zpz` | Main database service ID |
| **TIGER_SERVICE_ID** | `wuj5xa6zpz` | Primary service for MCP |
| **TIGER_DB_USER** | `tsdbadmin` | PostgreSQL admin user |
| **TIGER_DB_PASSWORD** | `ivx5ndpuyodbxw5w` | PostgreSQL admin password ⚠️ |
| **TIGER_DB_HOST** | `wuj5xa6zpz.a1lqw18o6u.tsdb.cloud.timescale.com` | Tiger Cloud host |
| **TIGER_DB_PORT** | `35548` | PostgreSQL port (non-standard) |
| **TIGER_DB_NAME** | `tsdb` | Database name |
| **TIGER_DB_SSLMODE** | `require` | SSL encryption required |
| **DATABASE_URL** | `postgres://tsdbadmin:ivx5ndpuyodbxw5w@...` | Full connection string |
| **TIGER_PUBLIC_KEY** | `01K9AGD3J4F904GTXZXMKPB1TC` | API public key |
| **TIGER_SECRET_KEY** | `62qoGWAmuwCjtKFzZd8iEMumagjPgheWywa9uDHAuacv9zNcvONIU2m2UHOqZ1CE` | API secret key ⚠️ |
| **TIGER_MCP_URL** | `http://mcp:9090` | MCP server endpoint (local Docker) |

### 🔀 Fork Services (Agents)

#### Fork A1 - Agent 1
```
TIGER_FORK_A1_SERVICE_ID=gwb579t287
TIGER_FORK_A1_SERVICE_NAME=afs-fork-agent-1
TIGER_FORK_A1_DATABASE_NAME=tsdb
TIGER_FORK_A1_USERNAME=tsdbadmin
TIGER_FORK_A1_PASSWORD=kee83tu9wbqwzzzz
TIGER_FORK_A1_SERVICE_URL=postgres://tsdbadmin:kee83tu9wbqwzzzz@gwb579t287.a1lqw18o6u.tsdb.cloud.timescale.com:34144/tsdb?sslmode=require
TIGER_FORK_A1_PORT=34144
TIGER_FORK_A1_HOST=gwb579t287.a1lqw18o6u.tsdb.cloud.timescale.com
```

#### Fork A2 - Agent 2
```
TIGER_FORK_A2_SERVICE_ID=mn4o89xewb
TIGER_FORK_A2_SERVICE_NAME=afs-fork-agent-2
TIGER_FORK_A2_DATABASE_NAME=tsdb
TIGER_FORK_A2_USERNAME=tsdbadmin
TIGER_FORK_A2_PASSWORD=qyczitkdgsj4zi1h
TIGER_FORK_A2_SERVICE_URL=postgres://tsdbadmin:qyczitkdgsj4zi1h@mn4o89xewb.a1lqw18o6u.tsdb.cloud.timescale.com:30080/tsdb?sslmode=require
TIGER_FORK_A2_PORT=30080
TIGER_FORK_A2_HOST=mn4o89xewb.a1lqw18o6u.tsdb.cloud.timescale.com
```

### 🔐 GCP Credentials (Google Cloud)

**Archivo**: `secrets/gcp_credentials.json`  
**Tipo**: Service Account JSON (23 KB)  
**Proyecto**: `divine-climate-476722-a2`  
**Usuario**: `vertex-express@divine-climate-476722-a2.iam.gserviceaccount.com`

**Modelos Vertex AI Configurados**:
- ✅ `gemini-2.5-pro` (CEREBRO - Reasoning)
- ✅ `gemini-2.5-flash` (OPERATIVO - Operations)
- ✅ `gemini-2.0-flash` (BULK - Bulk operations)

---

## 🔒 Manejo Seguro de GCP Credentials - Base64 Encoding

### Paso 1: Convertir JSON a Base64

```bash
# Desde Linux Debian:
cd /srv/afs-challenge
cat secrets/gcp_credentials.json | base64 -w 0 > /tmp/gcp_b64.txt
echo ""  # Nueva línea
cat /tmp/gcp_b64.txt
```

**Salida esperada**: String base64 de ~2000 caracteres comenzando con `ewogICJ...`

### Paso 2: Almacenar en Railway

En **Railway Dashboard → Project Settings → Variables**:

```
Variable Name: GCP_CREDENTIALS_BASE64
Value: [PASTE BASE64 STRING HERE]
```

### Paso 3: Decodificar en Railway Backend

En `backend/cmd/server/main.go` (o en tu init):

```go
package main

import (
    "encoding/base64"
    "os"
    "io/ioutil"
)

func init() {
    if b64 := os.Getenv("GCP_CREDENTIALS_BASE64"); b64 != "" {
        data, err := base64.StdEncoding.DecodeString(b64)
        if err != nil {
            panic("Invalid base64 GCP credentials: " + err.Error())
        }
        // Escribir a ubicación esperada
        err = ioutil.WriteFile("/app/gcp_credentials.json", data, 0600)
        if err != nil {
            panic("Failed to write GCP credentials: " + err.Error())
        }
        // Configurar variable de entorno
        os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/app/gcp_credentials.json")
    }
}
```

**Ventajas**:
- ✅ No expone credenciales en texto plano en logs
- ✅ Compatible con Railway dashboard UI
- ✅ Fácil rotación de credenciales
- ✅ Estándar de industria para secretos

---

## 🚢 RAILWAY - Backend Deployment

### 1. Crear Proyecto en Railway

1. Ir a https://railway.app
2. Click en **"New Project"**
3. Seleccionar **"Deploy from GitHub"**
4. Conectar a repositorio: `HCo-Innova/AgenticForkSquad`
5. Seleccionar rama: `main`
6. Click en **"Deploy"**

### 2. Configurar Backend Service

Railway debería detectar automáticamente el Dockerfile. Si no:

1. En Railway Project Dashboard
2. Click en tu deployment
3. Click en **"Settings"**
4. En **"Build Configuration"**:
   - **Builder**: `Dockerfile`
   - **Dockerfile Path**: `infrastructure/docker/backend/Dockerfile`
   - **Build Context**: `./backend` ← **IMPORTANTE**

### 3. Variables de Entorno - Railway Dashboard

Ve a: **Project → Variables** y añade estas variables:

#### 🌐 Tiger Cloud - Main Database
```
USE_TIGER_CLOUD=true
TIGER_PROJECT_ID=a1lqw18o6u
TIGER_MAIN_SERVICE=wuj5xa6zpz
TIGER_SERVICE_ID=wuj5xa6zpz
TIGER_DB_USER=tsdbadmin
TIGER_DB_PASSWORD=ivx5ndpuyodbxw5w
TIGER_DB_HOST=wuj5xa6zpz.a1lqw18o6u.tsdb.cloud.timescale.com
TIGER_DB_PORT=35548
TIGER_DB_NAME=tsdb
TIGER_DB_SSLMODE=require
DATABASE_URL=postgres://tsdbadmin:ivx5ndpuyodbxw5w@wuj5xa6zpz.a1lqw18o6u.tsdb.cloud.timescale.com:35548/tsdb?sslmode=require
```

#### 🔀 Tiger Cloud - Fork Services
```
TIGER_FORK_A1_SERVICE_ID=gwb579t287
TIGER_FORK_A1_PASSWORD=kee83tu9wbqwzzzz
TIGER_FORK_A1_HOST=gwb579t287.a1lqw18o6u.tsdb.cloud.timescale.com
TIGER_FORK_A1_PORT=34144
TIGER_FORK_A1_SERVICE_URL=postgres://tsdbadmin:kee83tu9wbqwzzzz@gwb579t287.a1lqw18o6u.tsdb.cloud.timescale.com:34144/tsdb?sslmode=require

TIGER_FORK_A2_SERVICE_ID=mn4o89xewb
TIGER_FORK_A2_PASSWORD=qyczitkdgsj4zi1h
TIGER_FORK_A2_HOST=mn4o89xewb.a1lqw18o6u.tsdb.cloud.timescale.com
TIGER_FORK_A2_PORT=30080
TIGER_FORK_A2_SERVICE_URL=postgres://tsdbadmin:qyczitkdgsj4zi1h@mn4o89xewb.a1lqw18o6u.tsdb.cloud.timescale.com:30080/tsdb?sslmode=require
```

#### 🔐 GCP Credentials (Base64 Encoded)
```
GCP_CREDENTIALS_BASE64=[PASTE_YOUR_BASE64_STRING_HERE]
VERTEX_PROJECT_ID=divine-climate-476722-a2
VERTEX_LOCATION=us-central1
GOOGLE_APPLICATION_CREDENTIALS=/app/gcp_credentials.json
```

#### 🤖 Gemini Models
```
GEMINI_CEREBRO_MODEL=gemini-2.5-pro
GEMINI_OPERATIVO_MODEL=gemini-2.5-flash
GEMINI_BULK_MODEL=gemini-2.0-flash
```

#### 🔧 Backend Configuration
```
PORT=8000
ENV=production
LOG_LEVEL=info
RUN_MIGRATIONS=true
TIGER_MCP_URL=http://mcp-service:9090
```

#### 📡 Frontend URLs (llenar después del deployment de Vercel)
```
VITE_API_URL=https://your-railway-app.railway.app/api/v1
VITE_WS_URL=wss://your-railway-app.railway.app/ws
```

### 4. Networking & Health Checks

**Railway URL**: Se genera automáticamente ej: `https://afs-backend-prod.railway.app`

**Health Check Endpoint**:
```bash
# Verificar desde CLI después del deployment:
curl -H "Authorization: Bearer YOUR_RAILWAY_TOKEN" \
  https://afs-backend-prod.railway.app/health
```

**Respuesta esperada**:
```json
{
  "status": "ok",
  "timestamp": "2025-11-08T12:34:56Z",
  "database": "connected",
  "redis": "connected",
  "tiger_cloud": "connected"
}
```

---

## 🎨 VERCEL - Frontend Deployment

### 1. Crear Proyecto en Vercel

1. Ir a https://vercel.com
2. Click en **"Add New..."** → **"Project"**
3. Seleccionar repositorio: `HCo-Innova/AgenticForkSquad`
4. Click en **"Import"**

### 2. Configurar Framework

En **Framework Preset**, seleccionar:
- **Framework**: `Vite`
- **Build Command**: `npm run build`
- **Output Directory**: `dist`
- **Install Command**: `npm ci`

### 3. Environment Variables - Vercel Dashboard

Ve a: **Settings → Environment Variables** y añade:

#### 📡 Backend API URLs (usar URL de Railway)
```
VITE_API_URL=https://your-railway-app.railway.app/api/v1
VITE_WS_URL=wss://your-railway-app.railway.app/ws
```

**Nota**: Reemplaza `your-railway-app` con la URL real de Railway  
Ej: `https://afs-backend-prod-a1b2.railway.app/api/v1`

#### ⚙️ Build Configuration
```
NODE_ENV=production
VITE_NODE_ENV=production
```

### 4. Deploy

1. Click en **"Deploy"**
2. Esperar a que Vercel compile y deploye (3-5 minutos)
3. Vercel genera URL automáticamente: `https://your-project.vercel.app`

### 5. Verificar Deployment

```bash
# Verificar que la app carga:
curl -I https://your-project.vercel.app

# Verificar que los env vars se inyectaron:
curl https://your-project.vercel.app/api/config
```

---

## ✅ HEALTH CHECKS - Verificar Conectividad

### 1. Backend Health (Railway)

```bash
# Desde terminal local:
RAILWAY_URL="https://your-railway-app.railway.app"

# Health check básico
curl -v "${RAILWAY_URL}/health"

# Verificar conectividad a Tiger Cloud
curl -v "${RAILWAY_URL}/api/v1/health"

# Verificar conexión a WebSocket
curl -i -N -H "Connection: Upgrade" \
  -H "Upgrade: websocket" \
  "${RAILWAY_URL}/ws"
```

**Respuestas esperadas**:

✅ **Backend Health**:
```json
{
  "status": "ok",
  "database": "connected",
  "redis": "connected",
  "tiger_cloud": {
    "status": "connected",
    "project": "a1lqw18o6u",
    "services": 3
  }
}
```

✅ **WebSocket Connection**:
```
HTTP/1.1 101 Switching Protocols
Connection: Upgrade
Upgrade: websocket
```

### 2. Frontend Health (Vercel)

```bash
VERCEL_URL="https://your-project.vercel.app"

# Load index
curl -I "${VERCEL_URL}"

# Verify env vars (check React bundle)
curl "${VERCEL_URL}/index.html" | grep -i "VITE_API_URL"

# Test API connection from browser console:
# fetch('${VITE_API_URL}/health').then(r => r.json()).then(console.log)
```

### 3. Tiger Cloud Connectivity Check

```bash
# Desde tu máquina local (requiere acceso a Tiger Cloud):
# Verificar que las credenciales funcionan

psql "${DATABASE_URL}" -c "SELECT version();"
psql "${TIGER_FORK_A1_SERVICE_URL}" -c "SELECT version();"
psql "${TIGER_FORK_A2_SERVICE_URL}" -c "SELECT version();"
```

**Respuesta esperada**: PostgreSQL version string

---

## 🔄 NETWORKING - Flujo de Datos

```
Usuario en Navegador
        ↓
    Vercel CDN (HTTPS)
        ↓
    Railway Backend (HTTPS/WSS)
        ↓
    Tiger Cloud PostgreSQL (TLS)
        ↓
    Vertex AI APIs (HTTPS)
```

**Puertos**:
- Vercel: `443` (HTTPS/WSS)
- Railway: `443` (HTTPS/WSS)
- Tiger Cloud: `35548` (PostgreSQL TLS)

**SSL/TLS**: ✅ End-to-end encryption

---

## 🚀 DEPLOYMENT CHECKLIST

### Pre-Deployment (Local)
- [ ] Todo código committeado a `main`
- [ ] `.env` NO está en git (verificado con pre-push check ✅)
- [ ] `secrets/` NO está en git (verificado con .gitignore)
- [ ] Docker builds exitosamente: `docker-compose build`
- [ ] Health endpoints responden localmente

### Railway Deployment
- [ ] Proyecto creado en Railway
- [ ] GitHub repo conectado
- [ ] Dockerfile path configurado
- [ ] Build context apunta a `./backend`
- [ ] Todas las env vars de Tiger Cloud añadidas
- [ ] GCP credentials en base64 configuradas
- [ ] Build completa exitosamente (verde)
- [ ] `/health` endpoint responde ✅
- [ ] WebSocket conecta exitosamente
- [ ] Tiger Cloud forks accesibles

### Vercel Deployment
- [ ] Proyecto creado en Vercel
- [ ] GitHub repo conectado
- [ ] Rama: `main`
- [ ] Framework: `Vite` detectado
- [ ] VITE_API_URL apunta a Railway URL correcta
- [ ] Build completa exitosamente
- [ ] App carga en navegador
- [ ] Conecta a backend exitosamente

### Post-Deployment
- [ ] Frontend se carga sin errores
- [ ] API endpoints responden desde frontend
- [ ] WebSocket conecta desde frontend
- [ ] Carga de datos desde Tiger Cloud
- [ ] Fork operations funcionan
- [ ] Gemini models responden

---

## 🆘 Troubleshooting

### Error: "Cannot reach Tiger Cloud"

```bash
# 1. Verificar DATABASE_URL
echo $DATABASE_URL

# 2. Verificar conexión desde Railway container
# (Via Railway logs)
psql $DATABASE_URL -c "SELECT 1;"

# 3. Si fail, verificar:
# - IP whitelist en Tiger Cloud console
# - Credenciales correctas
# - Port 35548 no bloqueado
```

### Error: "GCP credentials not found"

```go
// En Railway logs, buscar:
// "GOOGLE_APPLICATION_CREDENTIALS=/app/gcp_credentials.json"

// Verificar que base64 decodificó correctamente:
echo $GCP_CREDENTIALS_BASE64 | base64 -d > /tmp/test.json
cat /tmp/test.json | grep "project_id"
```

### Error: "WebSocket connection refused"

```bash
# 1. Verificar que backend escucha en /ws
curl -v wss://your-railway-app.railway.app/ws

# 2. Verificar CORS en Railway
# Headers: Access-Control-Allow-Origin debe tener Vercel URL

# 3. Verificar VITE_WS_URL en Vercel es correcta
# Debe ser: wss://your-railway-app.railway.app/ws
# NO: ws://your-railway-app.railway.app/ws (sin SSL)
```

### Error: "Migrations not running"

```bash
# En Railway backend logs, verificar:
# - RUN_MIGRATIONS=true está configurado
# - DATABASE_URL es válida
# - Migraciones no tienen syntax errors

# Correr manualmente:
goose -dir ./migrations postgres "$DATABASE_URL" up
```

---

## 📊 Monitoreo en Producción

### Railway Monitoring
- Ve a: **Project → Metrics**
- Monitora: CPU, Memory, Requests, Error Rate
- Configura alertas para CPU > 80%

### Vercel Analytics
- Ve a: **Project → Analytics**
- Monitora: Page load time, Core Web Vitals
- Configura alertas para Core Web Vitals

### Tiger Cloud Health
```bash
# Script de health check (guardar como cron job):
#!/bin/bash
for url in \
  "https://your-railway-app.railway.app/health" \
  "postgres://..." \
; do
  if ! curl -sf "${url}" > /dev/null 2>&1; then
    echo "ALERT: ${url} failed"
  fi
done
```

---

## 🎯 Resumen de URLs y Credenciales

| Componente | URL | Credenciales |
|-----------|-----|-------------|
| **GitHub Repo** | https://github.com/HCo-Innova/AgenticForkSquad | SSH: `id_github` |
| **Vercel Frontend** | https://your-project.vercel.app | OAuth (GitHub) |
| **Railway Backend** | https://your-railway-app.railway.app | Env vars en dashboard |
| **Tiger Cloud DB** | wuj5xa6zpz.a1lqw18o6u.tsdb.cloud.timescale.com:35548 | `tsdbadmin` / `ivx5ndpuyodbxw5w` |
| **GCP Service Account** | divine-climate-476722-a2 | `secrets/gcp_credentials.json` (base64) |

---

**✅ Documento completado**. Próximos pasos:
1. Ejecutar comando base64 para GCP credentials
2. Crear Railway project
3. Configurar variables
4. Crear Vercel project
5. Ejecutar health checks

