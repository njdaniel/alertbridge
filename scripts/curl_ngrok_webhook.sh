#!/usr/bin/env bash
set -euo pipefail

# Get the current ngrok public HTTPS URL
ngrok_url() {
    if ! command -v jq >/dev/null 2>&1; then
        echo "Error: 'jq' is required but not installed. Please install jq to parse ngrok API responses." >&2
        exit 1
    fi
    curl --silent http://127.0.0.1:4040/api/tunnels \
        | jq -r '.tunnels[] | select(.proto=="https") | .public_url'
}

# Send a test webhook to the ngrok-exposed AlertBridge endpoint
send_webhook() {
    local url
    url=$(ngrok_url)
    if [[ -z "$url" ]]; then
        echo "No ngrok HTTPS tunnel found. Is ngrok running?" >&2
        exit 1
    fi
    echo "Sending webhook to: $url"
    curl -X POST "$url/hook" \
        -H "Content-Type: application/json" \
        -d '{"bot":"test","symbol":"BTC/USD","side":"buy","qty":"0.0001"}'
}

# Example usage
send_webhook
