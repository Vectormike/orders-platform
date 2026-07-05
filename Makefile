SHELL := /bin/sh

.PHONY: help infra-up infra-down infra-reset infra-logs api relay worker-fulfillment worker-recovery worker-notifications e2e-local

help:
	@echo "Available targets:"
	@echo "  make infra-up            - start Postgres, Kafka, and Kafdrop"
	@echo "  make infra-down          - stop local infra services"
	@echo "  make infra-reset         - wipe and restart local infra"
	@echo "  make infra-logs          - stream local infra logs"
	@echo "  make api                 - run API service"
	@echo "  make relay               - run outbox relay"
	@echo "  make worker-fulfillment  - run fulfillment worker"
	@echo "  make worker-recovery     - run recovery worker"
	@echo "  make worker-notifications- run notifications worker"
	@echo "  make e2e-local           - print all commands for local E2E startup"

infra-up:
	docker compose up -d

infra-down:
	docker compose down

infra-reset:
	docker compose down -v
	docker compose up -d

infra-logs:
	docker compose logs -f

api:
	go run cmd/api/main.go

relay:
	go run cmd/outbox-relay/main.go

worker-fulfillment:
	go run cmd/worker-fulfillment/main.go

worker-recovery:
	go run cmd/worker-recovery/main.go

worker-notifications:
	go run cmd/worker-notifications/main.go

e2e-local:
	@echo "Run each command in a separate terminal:"
	@echo "  make api"
	@echo "  make relay"
	@echo "  make worker-fulfillment"
	@echo "  make worker-recovery"
	@echo "  make worker-notifications"
