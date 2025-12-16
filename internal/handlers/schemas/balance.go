package schemas

type BalanceResponse struct {
	Current   int64 `json:"current"`
	Withdrawn int64 `json:"withdrawn"`
}
