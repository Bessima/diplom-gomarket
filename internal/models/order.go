package models

type Order struct {
	ID      int         `json:"id"`
	UserID  int         `json:"user_id"`
	Amount  *string     `json:"amount,omitempty"`
	Accrual *string     `json:"accrual,omitempty"`
	Status  OrderStatus `json:"status"`
}

type OrderStatus string

const (
	NewStatus        OrderStatus = "NEW"
	InvalidStatus    OrderStatus = "INVALID"
	ProcessingStatus OrderStatus = "PROCESSING"
	ProcessedStatus  OrderStatus = "PROCESSED"
)
