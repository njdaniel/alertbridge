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

## Quickstart

1. Build and test:
   ```bash
   make build
   make test
   ```
2. Set required environment variables:
   ```bash
   export ALP_KEY=your_key
   export ALP_SECRET=your_secret
   ```
3. Run the application:
   ```bash
   ./alertbridge
   ```
4. By default the server listens on port `8080`. Override with the `PORT`
   environment variable if needed.

## Configuration

See [`docs/deployment.md`](docs/deployment.md) for Docker, Compose, and environment details (including how to test your deployment using `curl_domain_webhook.sh`).

Set `DEBUG_LOGGING=true` to log full webhook request bodies and client IPs when troubleshooting. Leave it unset or `false` in production to avoid storing sensitive data.

## Slack Integration

Configure either `SLACK_WEBHOOK_URL` for incoming webhooks or `SLACK_TOKEN` with `chat:write` permissions and `SLACK_CHANNEL` for OAuth-based posting. Optionally set `SLACK_NOTIFY` to control which events are sent (`success`, `failure`). When enabled, AlertBridge will post formatted messages to the specified Slack channel whenever orders succeed or fail. See [docs/slack.md](docs/slack.md) for full setup instructions.

## Webhook Format

See [`docs/webhook.md`](docs/webhook.md) for full details and examples.

## Metrics

Prometheus metrics are available at `/metrics`:

- `order_total{bot,side}`: Counter of processed orders

## Health Check

The `/healthz` endpoint returns `200 OK` and can be used for container liveness checks.

## Runbook
See [docs/runbook.md](docs/runbook.md) for deployment, monitoring, and rollback steps.

## Privacy

See [docs/privacy.md](docs/privacy.md) for information on how webhook payloads are logged, how long logs are kept, and how to request deletion.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

