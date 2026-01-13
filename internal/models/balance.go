package models

type Balance struct {
	UserID    int     `json:"-"`
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

func NewBalance(userID int) Balance {
	return Balance{
		UserID:    userID,
		Current:   0,
		Withdrawn: 0,
	}
}

func (balance *Balance) SetCurrent(value int32) {
	current := float32(value) / 100
	balance.Current = current
}

func (balance *Balance) SetWithdrawn(value int32) {
	withdrawn := float32(value) / 100
	balance.Withdrawn = withdrawn
}
