package models

import (
	"time"

	"gorm.io/gorm"
)

type Invoice struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"user_id"`
	ClientID  uint           `json:"client_id"`
	IssueDate string         `json:"issue_date" binding:"required"`
	DueDate   string         `json:"due_date" binding:"required"`
	Amount    float64        `json:"amount"`
	Status    string         `json:"status"` // "Pending", "Lunas", dll.
	Note      string         `json:"note"`
	Items     []InvoiceItem  `json:"items" gorm:"foreignKey:InvoiceID"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"` // optional, soft delete
}
