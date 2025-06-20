# Slack Integration

AlertBridge can send notifications to Slack whenever orders are processed. You can use either an incoming webhook URL or a bot token.

## Using an Incoming Webhook
1. Create a new [Incoming Webhook](https://api.slack.com/messaging/webhooks) in your Slack workspace.
2. Set the resulting URL in the `SLACK_WEBHOOK_URL` environment variable.

## Using a Bot Token
1. Create a Slack app and add the `chat:write` permission.
2. Install the app in your workspace and note the bot token (starts with `xoxb-`).
3. Set `SLACK_TOKEN` to this token and `SLACK_CHANNEL` to the channel ID where messages should be posted.

## Event Filtering
Use `SLACK_NOTIFY` to control which events send messages. Supported values:
- `success` – order created successfully.
- `failure` – risk check or order creation failed.

Separate multiple values with commas, e.g. `SLACK_NOTIFY=success,failure`.
