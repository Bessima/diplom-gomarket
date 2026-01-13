package models

import "time"

type Withdrawal struct {
	OrderID     string    `json:"order"`
	UserID      int       `json:"-"`
	Sum         float32   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

func (model *Withdrawal) SetSumInFloat(sumInt int32) {
	model.Sum = float32(sumInt) / 100
}
