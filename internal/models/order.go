package models

import "time"

type Order struct {
	ID         int         `json:"id"`
	UserID     int         `json:"user_id"`
	Accrual    *string     `json:"accrual,omitempty"`
	Status     OrderStatus `json:"status"`
	UploadedAt time.Time   `json:"uploaded_at"`
}

type OrderStatus string

const (
	NewStatus              OrderStatus = "NEW"
	InvalidStatus          OrderStatus = "INVALID"
	ProcessingStatus       OrderStatus = "PROCESSING"
	RegisterAcSystemStatus OrderStatus = "REGISTER"
	ProcessedStatus        OrderStatus = "PROCESSED"
)
