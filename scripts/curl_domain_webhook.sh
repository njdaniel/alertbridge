#!/usr/bin/env bash
set -euo pipefail

# Usage: ./curl_domain_webhook.sh example.com
DOMAIN=${1:?Usage: $0 <domain>}

# Send a test webhook to the production AlertBridge endpoint
send_webhook() {
    local url="https://$DOMAIN/hook"
    echo "Sending webhook to: $url"
    if ! curl --fail --silent --show-error -X POST "$url" \
        -H "Content-Type: application/json" \
        -d '{"bot":"test","symbol":"BTC/USD","side":"buy","qty":"0.0001"}'; then
        echo "Webhook request failed" >&2
        return 1
    fi
}

send_webhook
