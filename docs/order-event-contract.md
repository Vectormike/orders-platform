# Order Event Contract

This document defines the minimal event and status contract shared by API, outbox relay, fulfillment worker, and recovery worker.

## Events and topics

- `order.created` - produced by API through outbox when a new order is saved (internal persistence event).
- `order.processing` - produced by outbox relay after verification and published to Kafka as worker input.
- `order.fulfilled` - produced when fulfillment completes successfully.
- `order.incomplete` - produced when fulfillment fails but retry may still succeed.
- `order.failed` - produced when recovery decides no more retries should occur.

Topic mapping is one event per topic for now:

- topic `order.created` carries only `order.created` events, etc.

## Status transitions

Allowed transitions:

- `created -> processing`
- `created -> incomplete`
- `processing -> fulfilled`
- `processing -> incomplete`
- `incomplete -> processing`
- `incomplete -> failed`

Terminal statuses:

- `fulfilled`
- `failed`

## Worker responsibilities

- Fulfillment worker:
  - consumes `order.processing`
  - on success emits `order.fulfilled`
  - on retriable failure emits `order.incomplete`

- Recovery worker:
  - consumes `order.incomplete`
  - applies retry policy and backoff
  - emits `order.processing` to retry fulfillment, or emits `order.failed` after max retries
  - tracks `retry_count` and `last_failure_reason` on the order row

## Relay responsibility

- Outbox relay verifies outbox events before publish.
- For outbox `order.created`, relay verifies payload and remaps event to Kafka topic `order.processing`.
- Relay may publish `order.fulfilled`, `order.incomplete`, and `order.failed` directly without remapping.

## Fulfillment validation levels

Fulfillment worker validates in this order:

1. Format checks
   - `shipping_address_line` is present
   - `shipping_city` is present
   - `shipping_country_code` is ISO-2
   - `payment_token` is present and has minimum format
2. Business checks
   - country is currently one of `NG`, `US`, `GB`
   - amount is greater than zero
3. External verification hook
   - `ExternalVerifier` interface classifies provider outcomes:
     - retryable -> `order.incomplete`
     - terminal -> `order.failed`
   - current default implementation is token-based for local/dev behavior

Failure outcomes:

- Retryable failure -> emits `order.incomplete`
- Terminal failure -> emits `order.failed`
