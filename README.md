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
- `PORT`: Server port (default: 8080)
- `COOLDOWN_SEC`: Cooldown period in seconds (optional)
- `PROM_URL`: Prometheus base URL for PnL checks (optional)
- `PNL_MAX`: Maximum allowed PnL before blocking orders (optional)
- `PNL_MIN`: Minimum allowed PnL before blocking orders (optional)
- `TV_SECRET`: Shared secret for validating TradingView webhooks using the `X-TV-Signature` header (optional)
- `SLACK_WEBHOOK_URL`: Slack incoming webhook URL (optional)
- `SLACK_TOKEN`: Slack OAuth token with `chat:write` scope (optional)
- `SLACK_CHANNEL`: Channel ID used when `SLACK_TOKEN` is set
- `SLACK_NOTIFY`: Comma-separated events to send to Slack: `success` and/or `failure` (default: `success`)
- `DOMAIN`: Domain for Caddy HTTPS configuration

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

Use the provided Dockerfile together with the `.dockerignore` file to build the
image:

```bash
docker build -t alertbridge .
docker run -p 3000:3000 \
  -e ALP_KEY=your_key \
  -e ALP_SECRET=your_secret \
  alertbridge
```

## Docker Compose

This repository includes a `docker-compose.yml` for running AlertBridge together with Caddy and ngrok. Copy `.env.example` to `.env` (production) or `.env.local` (development) and fill in your Alpaca, ngrok, and DOMAIN values, then start the stack:
```bash
docker compose up
```

Services will be available on the following ports:

- **AlertBridge:** <http://localhost:3000>
- **Caddy:** https://<your-domain> (ports 80 and 443)
- **ngrok UI:** <http://localhost:4040>

### Local vs Production

Quick reference for running the stack in different environments. See
[`docs/environments.md`](docs/environments.md) for full details.

**Local development**

1. Copy `.env.example` to `.env.local` and set test credentials.
2. Start the stack:

   ```bash
   docker compose up
   ```

3. Expose the service externally with ngrok:

   ```bash
   ngrok http 8080
   ```

**Production**

1. Copy `.env.example` to `.env` and set real values (including `DOMAIN`).
2. Start the stack in detached mode:

   ```bash
   docker compose -f docker-compose.yml up -d
   ```

   Caddy terminates HTTPS and forwards traffic to AlertBridge.

## Slack Integration

Configure either `SLACK_WEBHOOK_URL` for incoming webhooks or `SLACK_TOKEN` with `chat:write` permissions and `SLACK_CHANNEL` for OAuth-based posting. Optionally set `SLACK_NOTIFY` to control which events are sent (`success`, `failure`). When enabled, AlertBridge will post formatted messages to the specified Slack channel whenever orders succeed or fail. See [docs/slack.md](docs/slack.md) for full setup instructions.

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


## Runbook
See [docs/runbook.md](docs/runbook.md) for deployment, monitoring, and rollback steps.


## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

