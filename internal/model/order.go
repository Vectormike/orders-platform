package model

import "time"

type Order struct {
	ID                  int64     `json:"id"`
	CustomerName        string    `json:"customer_name"`
	AmountCents         int64     `json:"amount_cents"`
	ShippingAddressLine string    `json:"shipping_address_line"`
	ShippingCity        string    `json:"shipping_city"`
	ShippingCountryCode string    `json:"shipping_country_code"`
	PaymentToken        string    `json:"payment_token"`
	Status              string    `json:"status"`
	CreatedAt           time.Time `json:"created_at"`
}
