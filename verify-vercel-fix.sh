#!/bin/bash

# VERCEL FIX CHECKLIST
# Ejecuta estos pasos en orden

echo "🔍 Vercel Frontend Deployment Fix Checklist"
echo "==========================================="
echo ""

echo "✅ Paso 1: Verificar archivos de configuración creados"
test -f vercel.json && echo "  ✓ vercel.json existe" || echo "  ✗ FALTA vercel.json"
test -f frontend/vite.config.ts && echo "  ✓ vite.config.ts actualizado" || echo "  ✗ PROBLEMA en vite.config.ts"
test -f frontend/package.json && echo "  ✓ package.json actualizado" || echo "  ✗ PROBLEMA en package.json"
test -f .env.production && echo "  ✓ .env.production existe" || echo "  ✗ FALTA .env.production"
echo ""

echo "✅ Paso 2: Verificar contenido de vercel.json"
echo "Expected:"
echo '  - buildCommand: "cd frontend && npm run build"'
echo '  - outputDirectory: "frontend/dist"'
echo '  - rewrites: [{ "source": "/(.*)", "destination": "/index.html" }]'
echo "Content:"
cat vercel.json | grep -E 'buildCommand|outputDirectory|source|destination' || echo "❌ No encontrado"
echo ""

echo "✅ Paso 3: Test build local"
echo "Ejecutando: cd frontend && npm run build"
cd frontend
npm run build

if [ -f dist/index.html ]; then
    echo "✓ dist/index.html existe - Size: $(du -sh dist/ | cut -f1)"
else
    echo "✗ FALTA dist/index.html"
fi
cd ..
echo ""

echo "✅ Paso 4: Instrucciones para Vercel Dashboard"
echo ""
echo "1. Ve a: https://vercel.com → Tu Proyecto"
echo "2. Settings → Environment Variables"
echo "3. Añade estas variables (reemplaza URLs con tu Railway URL):"
echo ""
echo "   VITE_API_URL = https://afs-backend-prod.railway.app/api/v1"
echo "   VITE_WS_URL = wss://afs-backend-prod.railway.app/ws"
echo "   NODE_ENV = production"
echo ""
echo "4. Click 'Save'"
echo "5. Deployments → Redeploy (último deployment)"
echo ""

echo "✅ Paso 5: Esperar deployment (3-5 minutos)"
echo "   Luego verifica: curl -I https://tu-vercel-url.vercel.app"
echo ""

echo "🎉 Si todo está correcto, deberías ver:"
echo "   HTTP/1.1 200 OK"
echo "   Content-Type: text/html"
echo ""
