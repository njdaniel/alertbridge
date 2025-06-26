# Docker & Deployment

## Docker Compose

This repository includes a `docker-compose.yml` for running AlertBridge together with Caddy and ngrok. For production, copy `.env.example` to `.env` (or `.env.production`) and place your Alpaca and DOMAIN values there. For local development with ngrok, copy `.env.local.example` to `.env.local`, set `NGROK_AUTHTOKEN`, and then start the stack:

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

1. Copy `.env.local.example` to `.env.local` only if you plan to use ngrok. The example file contains just `NGROK_AUTHTOKEN`.
2. Start the stack:

   ```bash
   docker compose up
   ```

3. Expose the service externally with ngrok:

   ```bash
   ngrok http 8080
   ```

#### Production

1. Copy `.env.example` to `.env` and set real values 
2. If using `docker-compose.prod.yml` for a Caddy-based deployment, copy
   `.env.production.example` to `.env.production` and update the image registry
   and domain values.
3. Start the stack in detached mode:


   ```bash
   docker compose --env-file .env.production \
     -f docker-compose.yml -f docker-compose.prod.yml up -d
   ```

   Caddy terminates HTTPS and forwards traffic to AlertBridge.

## Running from Source

You can also run AlertBridge directly without Docker. Build the binary and execute it with your Alpaca credentials set as environment variables:

```bash
go build -o alertbridge ./cmd/alertbridge
ALP_KEY=your_key ALP_SECRET=your_secret ./alertbridge
```

Alternatively run without building:

```bash
ALP_KEY=your_key ALP_SECRET=your_secret go run ./cmd/alertbridge
```
