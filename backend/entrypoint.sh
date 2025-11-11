#!/bin/sh
set -e

# Decode GCP credentials from base64 if provided
if [ -n "$GCP_CREDENTIALS_BASE64" ]; then
    echo "Decoding GCP credentials..."
    mkdir -p /app/secrets
    echo "$GCP_CREDENTIALS_BASE64" | base64 -d > /app/secrets/gcp_credentials.json
    export GOOGLE_APPLICATION_CREDENTIALS=/app/secrets/gcp_credentials.json
    echo "✅ GCP credentials decoded"
fi

# Set GCP_REGION if not provided
if [ -z "$GCP_REGION" ] && [ -n "$VERTEX_LOCATION" ]; then
    export GCP_REGION="$VERTEX_LOCATION"
fi

# Execute main application
exec "$@"
