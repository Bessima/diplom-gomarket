package models

import "time"

type Withdrawal struct {
	OrderID     int       `json:"order_id"`
	UserID      int       `json:"user_id"`
	Sum         *float32  `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}
