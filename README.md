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
# orders-platform

## Order Processing Flow

![Order Processing Flow](/Users/victorjonah/.cursor/projects/Users-victorjonah-Desktop-Projects-order-system/assets/image-e529c00e-02c3-4ed3-90d5-a831575279f9.png)

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
# orders-platform

## Order Processing Flow

![Order Processing Flow](https://mermaid.ai/live/edit#pako:eNrNV2tv6jgQ_StWvhRUSnlTokulNg0r1PaCAnsfKyRkEkOtOnHWdtqyVf_7OnYCIcnta7Xa5UtqzzkznuOZSfpsuNRDhmlw9GeEAhddYbhh0F8EQP5gJGgQ-SvE9DqETGAXhzAQwAKQA4tgFIii8WI6js2u753CEBftE8WmzENs6dJAMEpIWZDJbI_jiD1gF5WAnD2IoZByLCjbFnFXlzFuSrnYMMRBRTH4KY3Eij5Vi3jHvrn4maahUScMEVji-jqGXcP1PSzaRt9TH4-U3SN2so7IGhPilyo3UslkIK-m5OR9M-TSB1QKVY5T-4FXDbZOzs_lvZlgOpnNwalWR5vktjROLBNYDEGBJrGtkkg2iYmT2aENB2EkUsAsBjgpoCJPBX2eGh1pvLo0waX92_hrbm_8dWY7c3218sLq9XoNcAFFxIdHrnLmHVV_wVH3VUEPsYRiG6LhkXJTT3k1qc6WUOgN1X7ezcy-sa05CDfLgAq83laOtMdlgB4lV1cb9vI0a3J7O57vNhNlFHonRiJldtOKN-Veq9FMVPLSa1FVmHonEPtTFHg42FSqSY6A0Ud-iFVPE6wiTLxptCKY39mxEJUDCcDJuT5EPWTURZxLr9VDR9cmCDUfCBpid5jHg3u0HaZiFA98C9n9TAXW4smjVtPERt91ANn_PPJR4SgZ1EgWz0i3xIhRf7rDZLLaX8aopKRGuXtVFDCSo-P36dXF3M6g4mA3sm5I81g9WvrRBg-QYA8KTIM0BUgE-LbbBTxy43Np20FUHSQt5LSEky6XxVhCeauO9-RqCftj5XtAzZbw4T35PhaArtccJYMLEY6AgwTbwhVBYA0xiRh6d_44kD5DggT6jAAZdk3ONchp8B8rMUfMxwEkHxYixn-yCjTzfyBA8K6Jlbu5zAh71-DZMz82a5zSWVPwFqMc2f6OfkuOd3b9UkulyqntlMwb513zRqHigAF6EqqPLBrJd_VQBpAL-WkUr45BMztwctgvUgRr8s12fi5vL34sHXvujO3Z_s6c10tvP29VEe2iHh_XAIFcLJNaXuqUS_y-VaKZCNVMs-TSOB_-szwyjfAv5JB4rxZr_XNd5pR1mFPeXW811S8bp_Curiht5HePuItbEWRzAxWRTK_qK60lbUbN2DDsGaZgEaoZvmTBeGk8x7yFIe6QjxaGKf_0JG1hLIIXyZHfoH9Q6qc0RqPNnWGuoSyFmhGF8g2a_uux22UyQ8RUeRhms9FUTgzz2XiSy0673m62u41-o9fqnJ0NujVja5idTr0xaA_6je6g1-s1-93uS834S8Vt1M_6nYH89Vud1qDX6vVe_gbQtCE3)

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
