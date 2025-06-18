package models

import (
	"time"
)

type InvoiceItem struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	InvoiceID  uint      `json:"invoice_id"`
	ItemName   string    `json:"item_name"`
	Quantity   int       `json:"quantity"`
	UnitPrice  float64   `json:"unit_price"`
	TotalPrice float64   `json:"total_price"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
