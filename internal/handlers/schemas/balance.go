package schemas

type BalanceResponse struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

func NewBalanceResponse(current, withdrawn int32) BalanceResponse {
	return BalanceResponse{
		Current:   float32(current) / 100,
		Withdrawn: float32(withdrawn) / 100,
	}
}
