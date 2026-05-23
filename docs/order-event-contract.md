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

## Relay responsibility

- Outbox relay verifies outbox events before publish.
- For outbox `order.created`, relay verifies payload and remaps event to Kafka topic `order.processing`.
- Relay may publish `order.fulfilled`, `order.incomplete`, and `order.failed` directly without remapping.
