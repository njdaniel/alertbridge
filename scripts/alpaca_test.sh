#!/usr/bin/env bash
set -euo pipefail

# Load env vars if .env exists
if [[ -f .env ]]; then
  export $(grep -v '^#' .env | xargs)
fi

# Check that required vars are set
: "${ALP_KEY:?Need to set ALP_KEY}"
: "${ALP_SECRET:?Need to set ALP_SECRET}"
: "${ALP_BASE:="https://paper-api.alpaca.markets"}"

function get_account() {
  local response
  if ! response=$(curl --fail --silent --show-error -X GET "$ALP_BASE/v2/account" \
    -H "APCA-API-KEY-ID: $ALP_KEY" \
    -H "APCA-API-SECRET-KEY: $ALP_SECRET"); then
    echo "Failed to fetch account info" >&2
    return 1
  fi
  echo "$response" | jq
}

function place_order() {
  local symbol=$1 side=$2 qty=$3
  echo "Placing $side $qty of $symbol..."
  # Use tif=gtc for crypto, day for stocks
  local tif="day"
  if [[ $symbol == *"/"* ]]; then
    tif="gtc"
  fi
  if ! curl --fail --silent --show-error -X POST "$ALP_BASE/v2/orders" \
    -H "APCA-API-KEY-ID: $ALP_KEY" \
    -H "APCA-API-SECRET-KEY: $ALP_SECRET" \
    -H "Content-Type: application/json" \
    -d "$(jq -n \
          --arg symbol "$symbol" \
          --arg qty "$qty" \
          --arg side "$side" \
          --arg type "market" \
          --arg tif "$tif" \
          '{symbol: $symbol, qty: $qty, side: $side, type: $type, time_in_force: $tif}')"; then
    echo "Order request failed" >&2
    return 1
  fi
}

# Example usage
echo "=== Account Info ==="
get_account

# Uncomment to test a stock order:
# place_order "AAPL" "buy" "1"

# Uncomment to test a crypto order (ensure crypto enabled):
place_order "BTC/USD" "buy" "0.0001"

