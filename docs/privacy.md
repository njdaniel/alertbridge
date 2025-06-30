# Privacy & Data Retention

AlertBridge logs incoming webhook requests to aid debugging and auditing. Each log entry contains the remote IP address, user agent, and the full JSON payload that was sent to `/hook`.

## Log Storage

Logs are written to standard output. When running via Docker or another container runtime they are captured as container logs. The project itself does not store webhook payloads anywhere else.

## Retention

Container logs are rotated by the runtime. With Docker's default settings they are kept for about 30 days before being removed. If you ship the logs to an external system, retention will follow that system's policy.

## Requesting Deletion

To have a specific webhook payload removed from our logs before the retention period ends, open a GitHub issue with the timestamp and any relevant details. We will locate and delete the entry. If you operate your own instance, delete the corresponding logs from your logging system.
