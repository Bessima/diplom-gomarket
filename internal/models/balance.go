package models

type Balance struct {
	UserID      int     `json:"-"`
	Current     float32 `json:"current"`
	Withdrawing float32 `json:"withdrawing"`
}

func NewBalance(userID int) Balance {
	return Balance{
		UserID:      userID,
		Current:     0,
		Withdrawing: 0,
	}
}

func (balance *Balance) SetCurrent(value int32) {
	current := float32(value) / 100
	balance.Current = current
}

func (balance *Balance) SetWithdrawing(value int32) {
	withdrawing := float32(value) / 100
	balance.Withdrawing = withdrawing
}
