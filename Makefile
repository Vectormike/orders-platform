SHELL := /bin/sh

.PHONY: help api relay worker-fulfillment worker-recovery e2e-local

help:
	@echo "Available targets:"
	@echo "  make api                 - run API service"
	@echo "  make relay               - run outbox relay"
	@echo "  make worker-fulfillment  - run fulfillment worker"
	@echo "  make worker-recovery     - run recovery worker"
	@echo "  make e2e-local           - print all commands for local E2E startup"

api:
	go run cmd/api/main.go

relay:
	go run cmd/outbox-relay/main.go

worker-fulfillment:
	go run cmd/worker-fulfillment/main.go

worker-recovery:
	go run cmd/worker-recovery/main.go

e2e-local:
	@echo "Run each command in a separate terminal:"
	@echo "  make api"
	@echo "  make relay"
	@echo "  make worker-fulfillment"
	@echo "  make worker-recovery"
