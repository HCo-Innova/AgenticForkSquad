#!/usr/bin/env python3
import sys
import google.auth
import google.auth.transport.requests

SCOPE = ["https://www.googleapis.com/auth/cloud-platform"]

def generate_access_token():
    request = google.auth.transport.requests.Request()
    creds = None
    try:
        # Prefer ADC: env var GOOGLE_APPLICATION_CREDENTIALS or gcloud ADC
        creds, _ = google.auth.default(scopes=SCOPE)
    except Exception:
        # Fallback to local service account file
        creds, _ = google.auth.load_credentials_from_file(
            "./secrets/gcp_credentials.json", scopes=SCOPE
        )
    creds.refresh(request)
    print(creds.token)

if __name__ == "__main__":
    try:
        generate_access_token()
    except Exception as e:
        print(f"Error generating access token: {e}", file=sys.stderr)
        sys.exit(1)
