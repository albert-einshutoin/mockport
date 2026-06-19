# Line example

This minimal example uses the LINE-like adapter with fake local credentials.

```bash
docker build -t mockport:local -f docker/Dockerfile .
mockport run --config examples/line/mockport.yml
```

Use these values in the application under test:

```env
LINE_API_BASE_URL=http://localhost:43101/line
LINE_CHANNEL_ID=mockport_line_channel
LINE_CHANNEL_SECRET=mockport_line_secret
LINE_CHANNEL_TOKEN=mockport_line_channel_token
LINE_LIFF_ID=mockport-line-liff
LINE_MINI_DAPP_CLIENT_ID=mockport_line_mini_dapp_client
LINE_PAY_CHANNEL_ID=mockport_line_pay_channel
LINE_PAY_CHANNEL_SECRET=mockport_line_pay_secret
```

Smoke test:

```bash
curl -X POST "http://localhost:43101/line/v2/bot/message/push" \
  -d "to=C_MOCKPORT" \
  -d "messages=%5B%7B%22type%22%3A%22text%22%2C%22text%22%3A%22Mockport%22%7D%5D"
curl -X GET "http://localhost:43101/line/v2/bot/info"
curl http://localhost:43101/_mockport/report
```
