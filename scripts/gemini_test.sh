#!/usr/bin/env bash
set -euo pipefail

# Obtain an access token: use existing ACCESS_TOKEN or fallback to ADC via gcloud
if [ -z "${ACCESS_TOKEN:-}" ]; then
  if command -v gcloud >/dev/null 2>&1; then
    ACCESS_TOKEN=$(gcloud auth application-default print-access-token)
  else
    echo "ACCESS_TOKEN not set and gcloud not found. Export ACCESS_TOKEN or install gcloud." >&2
    exit 1
  fi
fi

curl -sS -X POST \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  "https://us-central1-aiplatform.googleapis.com/v1/projects/divine-climate-476722-a2/locations/us-central1/publishers/google/models/gemini-2.5-pro:generateContent" \
  -d '{
    "contents": [
      {"role": "user", "parts": [{"text": "¿Cuál es la capital de Canadá?"}]}
    ],
    "generationConfig": {"temperature": 0.2, "maxOutputTokens": 512}
  }'
