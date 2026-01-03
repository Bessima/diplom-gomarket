package schemas

import (
	"strconv"
	"time"
)

type WithdrawRequest struct {
	Order string  `json:"order" validate:"required"`
	Sum   float32 `json:"sum" validate:"required"`
}

func (req WithdrawRequest) GetOrderAsInt() (int64, error) {
	return strconv.ParseInt(req.Order, 10, 64)
}

func (req WithdrawRequest) GetSumAsInt() int {
	return int(req.Sum * 100)
}

type WithdrawResponse struct {
	Order       string    `json:"order" validate:"required"`
	Sum         float32   `json:"sum" validate:"required"`
	ProcessedAt time.Time `json:"processed_at"`
}
