# ============================================
# Stage 1: Builder
# ============================================
FROM golang:1.25-alpine AS builder

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

RUN apk add --no-cache git ca-certificates tzdata && \
    rm -rf /var/cache/apk/*

WORKDIR /build

# Dependencies
COPY backend/go.mod backend/go.sum ./
RUN go mod download && go mod verify

# Source
COPY backend/ .

# Build optimizado
RUN go build \
    -ldflags="-w -s -X main.Version=1.0.0" \
    -trimpath \
    -o app \
    ./cmd/api

# ============================================
# Stage 2: Production (Distroless base con libc)
# ============================================
FROM gcr.io/distroless/base-debian12:nonroot

LABEL maintainer="tu-email@ejemplo.com"
LABEL org.opencontainers.image.source="https://github.com/tuusuario/afs-challenge"

WORKDIR /app

# Copy binario
COPY --from=builder /build/app /app/app

# Run as nonroot user (UID 65532)
USER 65532:65532

EXPOSE 8000

ENTRYPOINT ["/app/app"]