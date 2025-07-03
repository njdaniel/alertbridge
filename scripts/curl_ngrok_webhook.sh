#!/usr/bin/env bash
set -euo pipefail

# Get the current ngrok public HTTPS URL
ngrok_url() {
    curl --fail --silent --show-error http://127.0.0.1:4040/api/tunnels \
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
    if ! curl --fail --silent --show-error -X POST "$url/hook" \
        -H "Content-Type: application/json" \
        -d '{"bot":"test","symbol":"BTC/USD","side":"buy","qty":"0.0001"}'; then
        echo "Webhook request failed" >&2
        return 1
    fi
}

# Example usage
send_webhook
