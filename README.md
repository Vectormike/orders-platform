# orders-platform

## Order Processing Flow

![Order Processing Flow](docs/order-processing-flow.png)

## Code Map

- API entry: `cmd/api/main.go`
- HTTP and service:
  - `internal/controller/order_controller.go`
  - `internal/service/order_service.go`
- Order and outbox transaction:
  - `internal/repository/order_repository.go`
- Relay transform and publish:
  - `internal/outbox/relay.go`
  - `internal/outbox/kafka_publisher.go`
  - `internal/ordercontract/contract.go`
- Fulfillment worker and validations:
  - `cmd/worker-fulfillment/main.go`
  - `internal/fulfillment/repository.go`
  - `internal/fulfillment/validation.go`
  - `internal/fulfillment/external_verifier.go`
- Recovery worker:
  - `cmd/worker-recovery/main.go`
  - `internal/recovery/repository.go`
- Schema support:
  - `internal/database/schema.go`

## Verification

- End-to-end runbook: `docs/end-to-end-verification-playbook.md`
