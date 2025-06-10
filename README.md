# AlertBridge

AlertBridge is a headless gateway that receives TradingView (or any bot) webhook alerts and translates them into authenticated Alpaca orders. It never contains trading strategy logic—it only validates, applies risk rules, logs, and forwards orders.

## Features

- Webhook endpoint for receiving trading alerts
- HMAC authentication for TradingView webhooks
- Risk management rules (cooldown periods, PnL checks)
- Prometheus metrics integration
- Graceful shutdown handling
- Containerized deployment

## Prerequisites

- Go 1.21 or later
- Docker (optional, for containerized deployment)
- Alpaca API credentials

## Configuration

Environment variables:

- `ALP_KEY`: Alpaca API key
- `ALP_SECRET`: Alpaca API secret
- `ALP_BASE`: Alpaca API base URL (default: https://paper-api.alpaca.markets)
- `PORT`: Server port (default: 3000)
- `TV_SECRET`: TradingView webhook secret (optional)
- `COOLDOWN_SEC`: Cooldown period in seconds (optional)
- `PROM_URL`: Prometheus base URL for PnL checks (optional)
- `PNL_MAX`: Maximum allowed PnL before blocking orders (optional)
- `PNL_MIN`: Minimum allowed PnL before blocking orders (optional)

## Building

```bash
# Build binary
make build

# Run tests
make test

# Build Docker image
make docker
```

## Testing

Unit tests cover HMAC verification, cooldown enforcement, and HTTP handler
responses. Execute the suite with:

```bash
make test
```

## Running

```bash
# Set required environment variables
export ALP_KEY=your_key
export ALP_SECRET=your_secret

# Run the application
./alertbridge
```

## Docker

```bash
# Build and run with Docker
docker build -t alertbridge .
docker run -p 3000:3000 \
  -e ALP_KEY=your_key \
  -e ALP_SECRET=your_secret \
  alertbridge
```

## Webhook Format

Send POST requests to `/hook` with the following JSON body:

```json
{
  "bot": "strategy1",
  "symbol": "BTCUSD",
  "side": "buy",
  "qty": "10",
  "ts": 1234567890
}
```

Important notes about the webhook format:
- `symbol` must be in Alpaca's format:
  - For stocks: Use the standard ticker (e.g., "AAPL", "MSFT")
  - For crypto: Use the combined format (e.g., "BTCUSD", "ETHUSD")
  - Do not use forward slashes (e.g., use "BTCUSD" not "BTC/USD")
- `side` must be either "buy" or "sell"
- `qty` can be a number or "all"
- `ts` is optional and should be Unix timestamp in milliseconds

If `TV_SECRET` is set, include the `X-TV-Signature` header with the HMAC-SHA256 signature of the request body.

## Metrics

Prometheus metrics are available at `/metrics`:

- `order_total{bot,side}`: Counter of processed orders

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

