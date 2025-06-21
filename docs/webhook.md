# Webhook Format

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

**Notes:**
- `symbol` must be in Alpaca's format:
  - For stocks: Use the standard ticker (e.g., "AAPL", "MSFT")
  - For crypto: Use the combined format (e.g., "BTC/USD", "ETH/USD")
  - Do use forward slashes (e.g., use "BTC/USD" not "BTCUSD")
- `side` must be either "buy" or "sell"
- `qty` can be a number or "all"
- `ts` is optional and should be Unix timestamp in milliseconds
- When `TV_SECRET` is set, include an `X-TV-Signature` header with the HMAC SHA256 of the request body
