# Runbook

This document explains how to deploy, monitor, and roll back AlertBridge in production environments.

## Deployment

1. **Build and test the binary**
   ```bash
   make build
   make test
   ```
2. **Build the Docker image**
   ```bash
   docker build -t alertbridge:latest .
   ```
3. **Deploy** the container to your infrastructure or run the provided `docker-compose.yml` stack:
   ```bash
   docker compose up -d
   ```

## Monitoring

- Prometheus metrics are exposed at `http://<host>:3000/metrics`. Use the provided `prometheus.yml` and Grafana dashboards to track request counts and errors.
- Review container logs regularly for failed webhook deliveries or risk rule violations.

## Rollback

1. Identify the previous working Docker image tag or git commit.
2. Redeploy using that tag:
   ```bash
   docker run -d --rm --name alertbridge_previous <image:tag>
   ```
3. If using Docker Compose, update the image tag in `docker-compose.yml` and run:
   ```bash
   docker compose up -d
   ```

