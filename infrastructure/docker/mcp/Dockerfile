FROM alpine:3.18

ARG TIGER_VERSION=0.15.0
# Example values: linux-amd64, linux-arm64, linux-armv7
ARG TIGER_ARCH=linux-amd64

RUN apk add --no-cache ca-certificates curl jq bash

WORKDIR /usr/local/bin

# Fetch asset URL from GitHub Releases API and download the matching binary.
# This avoids hardcoding filenames that may differ between releases.
RUN set -eux; \
    API_URL="https://api.github.com/repos/tigercloud/tiger/releases/tags/v${TIGER_VERSION}"; \
    echo "Fetching release metadata from ${API_URL}"; \
    ASSET_URL=$(curl -sSL "$API_URL" | jq -r --arg arch "$TIGER_ARCH" '.assets[] | select(.name | test($arch; "i")) | .browser_download_url' | head -n1); \
    if [ -z "${ASSET_URL}" ]; then echo "No matching asset found for arch=${TIGER_ARCH} in release v${TIGER_VERSION}"; exit 1; fi; \
    echo "Downloading $ASSET_URL"; \
    curl -fsSL -o tiger "$ASSET_URL"; \
    chmod +x tiger; \
    ls -la /usr/local/bin/tiger

EXPOSE 9090

CMD ["/usr/local/bin/tiger", "mcp", "start", "http", "--host", "0.0.0.0", "--port", "9090"]
