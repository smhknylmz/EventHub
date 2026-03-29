# EventHub

Event-Driven Notification System that processes and delivers messages through multiple channels (sms, email, push) with reliable delivery, retry logic, and real-time status tracking

## Architecture

```
Client → Http (Echo) → PostgreSql (persistence)
                    → Redis Streams (queue)
                    → Worker (dispatcher + processor)
                    → Webhook Provider (delivery)
                    → Retry Poller (failed → re-queue)
```

### Key Components

- **Api Layer** — Crud endpoints with validation, pagination, idempotency
- **Queue** — Redis Streams with 9 priority-weighted channels (3 channels × 3 priorities)
- **Worker** — Dispatcher reads from streams, processor delivers via webhook
- **Rate Limiter** — Redis increment based fixed window (100 msg/sec/channel)
- **Retry** — Exponential backoff with configurable max retries and dead letter
- **Observability** — Prometheus metrics, structured logging with correlation IDs

## Quick Start

```bash
docker-compose up
```

The api will be available at `http://localhost:8080`

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | required | PostgreSql connection string |
| `REDIS_URL` | required | Redis connection string |
| `WEBHOOK_BASE_URL` | required | Webhook delivery url |
| `SERVER_PORT` | 8080 | Http server port |
| `LOG_LEVEL` | info | Log level (debug, info, warn, error) |
| `MAX_RETRIES` | 5 | Max delivery retry attempts |
| `BACKOFF_BASE` | 1s | Base duration for exponential backoff |
| `RETRY_POLL_INTERVAL` | 5s | Retry poller check interval |
| `IDEMPOTENCY_TTL` | 24h | Idempotency key ttl |

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/notifications` | Create notification |
| `POST` | `/api/v1/notifications/batch` | Create batch (max 1000) |
| `GET` | `/api/v1/notifications/:id` | Get by ID |
| `GET` | `/api/v1/notifications` | List with filters |
| `PATCH` | `/api/v1/notifications/:id/cancel` | Cancel pending |
| `GET` | `/health` | Health check |
| `GET` | `/metrics` | Prometheus metrics |
| `GET` | `/swagger/index.html` | API documentation |

### Examples

**Create notification:**
```bash
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "Content-Type: application/json" \
  -d '{"recipient": "+905551234567", "channel": "sms", "content": "Hello", "priority": "high"}'
```

**Create batch:**
```bash
curl -X POST http://localhost:8080/api/v1/notifications/batch \
  -H "Content-Type: application/json" \
  -d '{"notifications": [
    {"recipient": "user@email.com", "channel": "email", "content": "Test", "priority": "normal"},
    {"recipient": "+905559999999", "channel": "sms", "content": "Test", "priority": "high"}
  ]}'
```

**List with filters:**
```bash
curl "http://localhost:8080/api/v1/notifications?status=delivered&channel=sms&page=1&pageSize=20"
```

**Idempotent request:**
```bash
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: unique-key-123" \
  -d '{"recipient": "+905551234567", "channel": "sms", "content": "Hello"}'
```

## Running Tests

```bash
go test ./...
```