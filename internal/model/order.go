package model

import "time"

type Order struct {
	ID           int64     `json:"id"`
	CustomerName string    `json:"customer_name"`
	AmountCents  int64     `json:"amount_cents"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}
