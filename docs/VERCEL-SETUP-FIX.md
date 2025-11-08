# 🚀 Vercel Deployment Guide - Frontend Setup

## Problema Identificado (404: NOT_FOUND)

El error 404 en Vercel ocurre porque:
1. ❌ No había `vercel.json` para configurar build/output correctamente
2. ❌ Faltaban rewrites para SPA React
3. ❌ Vite no tenía configuración de build optimizada

## ✅ Soluciones Implementadas

### 1. **vercel.json Creado**
- ✅ Define `buildCommand`: `cd frontend && npm run build`
- ✅ Define `outputDirectory`: `frontend/dist`
- ✅ Define `installCommand`: `cd frontend && npm install`
- ✅ **Rewrites para SPA**: Redirecciona todas las rutas a `/index.html`

### 2. **vite.config.ts Actualizado**
```typescript
build: {
  outDir: 'dist',           // Output directory
  sourcemap: false,         // No sourcemaps en prod
  minify: 'terser',        // Minificación
  target: 'ES2020'         // Target moderno
}
```

### 3. **package.json Actualizado**
```json
"build": "tsc && vite build"  // TypeScript check + Vite build
```

### 4. **.env.production Creado**
- Contiene variables de ejemplo para testing local
- Debe ser reemplazado en Vercel Dashboard

---

## 📋 Pasos para Configurar en Vercel

### Paso 1: Acceder al Dashboard de Vercel
1. Ve a https://vercel.com
2. Selecciona tu proyecto `agentic-fork-squad-vercel`

### Paso 2: Configurar Environment Variables
Ve a: **Settings → Environment Variables**

Añade estas variables (reemplaza los valores con tu Railway URL):

```
VITE_API_URL = https://afs-backend-prod.railway.app/api/v1
VITE_WS_URL = wss://afs-backend-prod.railway.app/ws
NODE_ENV = production
```

**Nota**: Reemplaza `afs-backend-prod.railway.app` con tu URL real de Railway

### Paso 3: Triggear New Deployment
1. Ve a **Deployments**
2. Click en el último deployment
3. Click en **Redeploy** o **Promote to Production**
4. O push un cambio a `main` branch

### Paso 4: Verificar Deployment

```bash
# Espera 3-5 minutos por el build
# Accede a tu URL de Vercel

# Expected: Página carga correctamente
# Check en browser console: Sin errores 404 o CORS
```

---

## 🧪 Testing Local Vercel Build

Antes de deployar, puedes probar localmente:

```bash
# 1. Build el proyecto
cd frontend
npm run build

# 2. Verifica que frontend/dist existe y tiene contenido
ls -la dist/

# 3. Ver el contenido:
# dist/index.html              # Archivo HTML principal
# dist/assets/                 # JavaScript, CSS, etc.
```

---

## 🔍 Troubleshooting

### ❌ Aún sale 404 después de redeploy

**Causa**: Cache de Vercel

**Solución**:
1. Ve a **Settings → Git**
2. Disable de redeploy
3. Wait 5 minutos
4. Click "Redeploy"

### ❌ Variables de entorno no se inyectan

**Causa**: Variables no están en producción

**Solución**:
1. Ve a **Settings → Environment Variables**
2. Asegúrate de que están en **Production** environment
3. Redeploy

### ❌ API calls fallan (CORS)

**Causa**: `VITE_API_URL` es incorrecto

**Solución**:
1. Abre browser console (F12)
2. Verifica que `fetch('${VITE_API_URL}/health')` funciona
3. Actualiza `VITE_API_URL` en Vercel dashboard
4. Redeploy

---

## 📊 Vercel Build Output (Esperado)

Después de estos cambios, deberías ver en los logs:

```
✓ Build Completed
✓ Deploying outputs...
✓ Deployment completed
✓ Creating build cache...
```

Y la página debe cargar sin 404.

---

## 🔗 Próximas Validaciones

1. **Frontend carga** → `curl https://your-vercel-url.vercel.app`
2. **Backend conecta** → Check en console: Network tab
3. **WebSocket funciona** → Network → WS tab (debe haber conexión)
4. **API responses** → Console Network tab (status 200)

---

## 📝 Documentación Relacionada

- [DEPLOYMENT-VERCEL-RAILWAY.md](../DEPLOYMENT-VERCEL-RAILWAY.md)
- [11-DEPLOYMENT-STRATEGY.md](../11-DEPLOYMENT-STRATEGY.md)
- [Frontend Components](./09-FRONTEND-COMPONENTS.md)

---

**Status**: ✅ Ready for deployment  
**Last Updated**: November 8, 2025
