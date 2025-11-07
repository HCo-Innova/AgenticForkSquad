# ============================================
# Stage 1: Builder
# ============================================
FROM node:22-alpine AS builder

WORKDIR /app

# Dependencies
COPY frontend/package*.json ./
RUN npm ci && npm cache clean --force

# Build
COPY frontend/ ./
RUN npm run build

# ============================================
# Stage 2: Production (Caddy)
# ============================================
FROM caddy:2.8-alpine

LABEL maintainer="tu-email@ejemplo.com"

# Copy assets
COPY --from=builder /app/dist /usr/share/caddy

# Copy Caddyfile
COPY infrastructure/docker/frontend/Caddyfile /etc/caddy/Caddyfile

# Non-root (caddy user ya existe)
USER caddy

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost/health || exit 1

CMD ["caddy", "run", "--config", "/etc/caddy/Caddyfile", "--adapter", "caddyfile"]