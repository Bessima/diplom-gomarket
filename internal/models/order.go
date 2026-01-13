package models

import "time"

type Order struct {
	ID         string      `json:"number"`
	UserID     int         `json:"-"`
	Accrual    *float32    `json:"accrual,omitempty"`
	Status     OrderStatus `json:"status"`
	UploadedAt time.Time   `json:"uploaded_at"`
}

func (order *Order) SetAccrualAsFloat(accrualInt int32) {
	if accrualInt == 0 {
		return
	}
	val := float32(accrualInt) / 100
	order.Accrual = &val
}

type OrderStatus string

const (
	NewStatus              OrderStatus = "NEW"
	InvalidStatus          OrderStatus = "INVALID"
	ProcessingStatus       OrderStatus = "PROCESSING"
	RegisterAcSystemStatus OrderStatus = "REGISTER"
	ProcessedStatus        OrderStatus = "PROCESSED"
)
