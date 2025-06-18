package models

import (
	"time"

	"gorm.io/gorm"
)

type Client struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `json:"user_id"`
	Nama      string         `json:"name"`
	Email     string         `json:"email"`
	NoTlp     string         `json:"no_tlp"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
