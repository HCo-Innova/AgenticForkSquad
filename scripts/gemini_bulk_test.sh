#!/usr/bin/env bash
set -euo pipefail

# ACCESS_TOKEN fallback via ADC (gcloud)
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
  "https://us-central1-aiplatform.googleapis.com/v1/projects/divine-climate-476722-a2/locations/us-central1/publishers/google/models/gemini-2.0-flash:generateContent" \
  -d '{
    "contents": [
      {"role": "user", "parts": [{"text": "Dame un dato curioso sobre la historia de Linux."}]}
    ],
    "generationConfig": {"temperature": 0.7, "maxOutputTokens": 512}
  }'