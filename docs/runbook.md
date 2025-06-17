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

## Base Image Digest Updates

To pin the Docker base image to the latest Distroless digest:

1. Fetch the current digest:
   ```bash
   curl -s https://gcr.io/v2/distroless/static-debian11/manifests/latest \
     | grep -o 'sha256:[0-9a-f]\{64\}' | head -n 1
   ```
2. Update the `FROM` line in the `Dockerfile` with the retrieved digest.
3. Rebuild and redeploy the image.


## SBOM and Vulnerability Report Updates

To regenerate the software bill of materials and vulnerability report:

1. Install the tooling if not already available:
   ```bash
   go install golang.org/x/vuln/cmd/govulncheck@v1.1.4
   curl -sSfL https://raw.githubusercontent.com/anchore/syft/v0.84.0/install.sh | sh -s -- -b /usr/local/bin
   ```
2. Run the scanners from the repository root:
   ```bash
   govulncheck ./... 2>&1 | tee artifacts/govulncheck.txt
   syft . -o spdx-json > artifacts/sbom.spdx
   ```
3. Commit the updated files in the `artifacts/` directory.
