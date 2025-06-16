# AlertBridge

AlertBridge is a headless gateway that receives TradingView (or any bot) webhook alerts and translates them into authenticated Alpaca orders. It never contains trading strategy logic—it only validates, applies risk rules, logs, and forwards orders.

## Features

- Webhook endpoint for receiving trading alerts
- Risk management rules (cooldown periods, PnL checks)
- Prometheus metrics integration
- Health check endpoint (`/healthz`)
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
- `COOLDOWN_SEC`: Cooldown period in seconds (optional)
- `PROM_URL`: Prometheus base URL for PnL checks (optional)
- `PNL_MAX`: Maximum allowed PnL before blocking orders (optional)
- `PNL_MIN`: Minimum allowed PnL before blocking orders (optional)
- `TV_SECRET`: Shared secret for validating TradingView webhooks using the `X-TV-Signature` header (optional)
- `GF_SECURITY_ADMIN_PASSWORD`: Grafana admin password when using Docker Compose

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

Unit tests cover cooldown enforcement and HTTP handler responses. Execute the suite with:

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

## Docker Compose

This repository includes a `docker-compose.yml` for running AlertBridge together
with Prometheus, Grafana, and ngrok. Create a `.env` file based on
`.env.example` with your Alpaca, ngrok, and Grafana credentials, then start the stack:

```bash
docker compose up
```

Services will be available on the following ports:

- **AlertBridge:** <http://localhost:3000>
- **Prometheus:** <http://localhost:9090>
- **Grafana:** <http://localhost:3001> (login with `GF_SECURITY_ADMIN_PASSWORD`)
- **ngrok UI:** <http://localhost:4040>

## Webhook Format

Send POST requests to `/hook` with the following JSON body:

```json
{
  "bot": "strategy1",
  "symbol": "BTC/USD",
  "side": "buy",
  "qty": "10",
  "ts": 1234567890
}
```

Important notes about the webhook format:
- `symbol` must be in Alpaca's format:
  - For stocks: Use the standard ticker (e.g., "AAPL", "MSFT")
  - For crypto: Use the combined format (e.g., "BTC/USD", "ETH/USD")
  - Do use forward slashes (e.g., use "BTC/USD" not "BTCUSD")
- `side` must be either "buy" or "sell"
- `qty` can be a number or "all"
- `ts` is optional and should be Unix timestamp in milliseconds
- When `TV_SECRET` is set, include an `X-TV-Signature` header with the HMAC SHA256 of the request body

## Metrics

Prometheus metrics are available at `/metrics`:

- `order_total{bot,side}`: Counter of processed orders

## Health Check

The `/healthz` endpoint returns `200 OK` and can be used for container
liveness checks.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

