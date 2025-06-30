# Environment Setup

This guide explains how to configure AlertBridge for local development and for production deployments. Use the provided example files to keep sensitive values out of version control.

## .env.local vs .env

1. Copy `.env.local.example` to `.env.local` if you want to run ngrok locally. The file now only provides `NGROK_AUTHTOKEN`.
2. Copy `.env.example` to `.env` (or `.env.production`) for production deployments. Set your Alpaca credentials and include your production domain in the `DOMAIN` variable.
3. Both `.env` and `.env.local` are ignored by Git (see `.gitignore`) so secrets won't be committed accidentally.

Load either file when running Docker Compose with the `env_file` directive. `docker-compose.yml` uses `.env` by default while `docker-compose.override.yml` loads `.env.local` for development.

## Local Testing with ngrok

For quick testing against webhooks from external services, expose the local server using ngrok:

```bash
# Start the compose stack
docker compose up

# In another terminal, launch ngrok on port 8080
ngrok http 8080
```

If you configured `NGROK_AUTHTOKEN` in `.env.local`, Docker Compose will start an ngrok container automatically.

## Production with Caddy

Production deployments typically run behind Caddy to provide HTTPS. Set the `DOMAIN` variable in `.env` and start the stack in detached mode:

```bash
docker compose -f docker-compose.yml up -d
```

Caddy reads `DOMAIN` from the environment and proxies HTTPS traffic to the AlertBridge container. See `Caddyfile` for the minimal reverse proxy configuration.

