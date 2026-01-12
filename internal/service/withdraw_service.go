package service

import (
	"errors"
	"github.com/Bessima/diplom-gomarket/internal/handlers/schemas"
	"github.com/Bessima/diplom-gomarket/internal/models"
	"github.com/Bessima/diplom-gomarket/internal/repository"
	"sync"
)

type BalanceRepositoryI interface {
	SetWithdrawForUserID(userID int, withdraw int) error
}

type WithdrawService struct {
	BalanceRepository  BalanceRepositoryI
	WithdrawRepository repository.WithdrawStorageRepositoryI
	mu                 sync.Mutex
}

func NewWithdrawService(withdrawRep repository.WithdrawStorageRepositoryI, balanceRep BalanceRepositoryI) *WithdrawService {
	return &WithdrawService{BalanceRepository: balanceRep, WithdrawRepository: withdrawRep}
}

func (service *WithdrawService) Set(user *models.User, withdrawRequest schemas.WithdrawRequest) error {
	service.mu.Lock()
	defer service.mu.Unlock()

	orderID, err := withdrawRequest.GetOrderAsInt()
	if err != nil {
		return errors.New("can't parse number of order")
	}
	withdrawInt := withdrawRequest.GetSumAsInt()
	err = service.WithdrawRepository.Create(user.ID, orderID, withdrawInt)
	if err != nil {
		return err
	}

	err = service.BalanceRepository.SetWithdrawForUserID(user.ID, withdrawInt)

	return err

}
