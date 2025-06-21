# Docker & Deployment

## Docker Compose

This repository includes a `docker-compose.yml` for running AlertBridge together with Caddy and ngrok. Copy `.env.example` to `.env` (production) or `.env.local` (development) and fill in your Alpaca, ngrok, and DOMAIN values, then start the stack:

```bash
docker compose up
```

Services will be available on the following ports:

- **AlertBridge:** <http://localhost:8080>
- **Caddy:** https://<your-domain> (ports 80 and 443)
- **ngrok UI:** <http://localhost:4040>

### Local vs Production

Quick reference for running the stack in different environments.

#### Local development

1. Copy `.env.local.example` to `.env.local` and set test credentials.
2. Start the stack:

   ```bash
   docker compose up
   ```

3. Expose the service externally with ngrok:

   ```bash
   ngrok http 8080
   ```

#### Production

1. Copy `.env.example` to `.env` and set real values (including `DOMAIN`).
2. Start the stack in detached mode:

   ```bash
   docker compose -f docker-compose.yml up -d
   ```

   Caddy terminates HTTPS and forwards traffic to AlertBridge.
