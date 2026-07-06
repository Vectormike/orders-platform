# Orders Platform

Build an order pipeline that is simple, observable, and event-driven.

You send an order once, and the system handles the rest:
- persist order + event atomically
- relay events from Postgres outbox to Kafka
- run fulfillment + recovery workers
- notify customers through WhatsApp (Twilio/Meta) or stdout

## What This Project Is

This repo is a Go-based order system using:
- `Gin` for HTTP API
- `Postgres` for source of truth
- `Outbox pattern` for reliable event publishing
- `Kafka` for async workflow
- workers for fulfillment, recovery, and notifications

## Quick Start

1. Copy env template:

```bash
cp .env.example .env
```

2. Start infra:

```bash
make infra-up
```

3. Run services in separate terminals:

```bash
make api
make relay
make worker-fulfillment
make worker-recovery
make worker-notifications
```

4. Open tools:
- API health: `http://127.0.0.1:8080/health`
- Kafka UI (Kafdrop): `http://localhost:19000`
- Postgres host port: `5433`

## Architecture At A Glance

### Overview

![Architecture Overview](docs/architecture-overview.png)

### Order Flow

![Order Processing Flow](docs/order-processing-flow.png)

## The Story Of One Order

1. API writes `orders` + `outbox` in one transaction.
2. Outbox relay picks pending events and publishes to Kafka.
3. Fulfillment worker validates and emits:
   - `order.fulfilled`, or
   - `order.incomplete`, or
   - `order.failed`
4. Recovery worker retries incomplete orders until max retries, then emits `order.failed`.
5. Notifications worker consumes fulfilled/incomplete/failed events and sends customer updates.

## Where To Look In Code

- API entry: `cmd/api/main.go`
- Router/controller/service: `internal/http/router/router.go`, `internal/controller/order_controller.go`, `internal/service/order_service.go`
- Order write + outbox tx: `internal/repository/order_repository.go`
- Outbox relay: `cmd/outbox-relay/main.go`, `internal/outbox/relay.go`
- Event contract: `internal/ordercontract/contract.go`
- Fulfillment worker: `cmd/worker-fulfillment/main.go`, `internal/fulfillment/repository.go`
- Recovery worker: `cmd/worker-recovery/main.go`, `internal/recovery/repository.go`
- Notifications worker: `cmd/worker-notifications/main.go`, `internal/notification/processor.go`
- Notification senders:
  - Twilio WhatsApp: `internal/notification/whatsapp_twilio_sender.go`
  - Meta WhatsApp: `internal/notification/whatsapp_webhook_sender.go`
  - Stdout: `internal/notification/stdout_sender.go`
- DB schema bootstrap: `internal/database/schema.go`

## Notifications Setup

### Local only (no external provider)

Use stdout:
- `NOTIFICATION_CHANNEL=stdout`

### WhatsApp delivery

Required basics:
- `NOTIFICATION_CHANNEL=whatsapp`
- `NOTIFICATION_RECIPIENT=+<customer_e164_phone>`
- `WHATSAPP_PROVIDER=twilio` or `meta`

Twilio (default):
- `TWILIO_ACCOUNT_SID=<twilio_account_sid>`
- `TWILIO_AUTH_TOKEN=<twilio_auth_token>`
- `TWILIO_WHATSAPP_FROM=whatsapp:+14155238886`

Meta Cloud API:
- `WHATSAPP_ACCESS_TOKEN=<meta_cloud_api_token>`
- `WHATSAPP_PHONE_NUMBER_ID=<meta_phone_number_id>`
- optional `WHATSAPP_API_VERSION=v21.0`

OpenAI message generation (optional):
- `OPENAI_API_KEY=<openai_api_key>`
- `OPENAI_MODEL=gpt-4.1-mini`
- `OPENAI_RESPONSES_ENDPOINT=https://api.openai.com/v1/responses`

If OpenAI is unavailable, the worker falls back to template messages.

## Testing And Verification

- End-to-end playbook: `docs/end-to-end-verification-playbook.md`
- Postman collection: `docs/postman/order-system.postman_collection.json`
- Env template: `.env.example`

If you just want a confidence check, run Scenario A/B/C from the playbook and compare final `orders` + `outbox` rows.
