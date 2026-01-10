package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/Bessima/diplom-gomarket/internal/customerror"
	"github.com/Bessima/diplom-gomarket/internal/handlers/schemas"
	"github.com/Bessima/diplom-gomarket/internal/middlewares/logger"
	"github.com/Bessima/diplom-gomarket/internal/repository"
	"github.com/Bessima/diplom-gomarket/internal/service"
	"io"
	"net/http"
)

type WithdrawHandler struct {
	WithdrawRepository repository.WithdrawStorageRepositoryI
	OrderRepository    repository.OrderStorageRepositoryI
	BalanceRepository  *repository.BalanceRepository
}

func NewWithdrawHandler(withdrawStorage repository.WithdrawStorageRepositoryI, orderStorage repository.OrderStorageRepositoryI, balanceRepository *repository.BalanceRepository) *WithdrawHandler {

	return &WithdrawHandler{
		WithdrawRepository: withdrawStorage,
		OrderRepository:    orderStorage,
		BalanceRepository:  balanceRepository,
	}
}

func (h *WithdrawHandler) Add(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "can't read body", http.StatusBadRequest)
		logger.Log.Error(err.Error())
		return
	}
	var body schemas.WithdrawRequest
	err = json.Unmarshal(bodyBytes, &body)
	if err != nil {
		http.Error(w, "can't parse body", http.StatusBadRequest)
		logger.Log.Error(err.Error())
		return
	}

	if !CheckLuhn(body.Order) {
		http.Error(w, "invalid order number", http.StatusUnprocessableEntity)
		logger.Log.Error(fmt.Sprintf("invalid order number: %s", body.Order))
		return
	}

	user := GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "user was not got", http.StatusBadRequest)
		logger.Log.Error("user was not got")
		return
	}

	balance, err := h.BalanceRepository.GetBalanceUserID(user.ID)
	if err != nil {
		http.Error(w, "can't get balance of user", http.StatusBadRequest)
		return
	}

	if body.Sum > balance.Current {
		http.Error(w, "There are insufficient funds", http.StatusPaymentRequired)
		return
	}

	withdrawService := service.NewWithdrawService(h.WithdrawRepository, h.BalanceRepository)
	err = withdrawService.Set(user, body)

	if err != nil {
		if customErr, ok := err.(customerror.CustomError); ok {
			http.Error(w, customErr.Error(), customErr.GetHTTPCode())
			logger.Log.Warn(customErr.Error())
			return
		}
		errWithMessage := fmt.Sprintf("error while setting withdraw `%v` for user %d", err, user.ID)
		http.Error(w, "withdraw was not installed for user", http.StatusInternalServerError)
		logger.Log.Warn(errWithMessage)
		return
	}

	w.Write([]byte("Withdraw successfully added!"))
	w.WriteHeader(http.StatusOK)

}

func (h *WithdrawHandler) GetList(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "user was not got", http.StatusBadRequest)
		logger.Log.Error("user was not got")
		return
	}

	withdrawals, err := h.WithdrawRepository.GetListByUserID(user.ID)
	if err != nil {
		http.Error(w, "withdrawals were not found", http.StatusInternalServerError)
		logger.Log.Error(fmt.Sprintf("withdrawals were not found, error: %v", err))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(withdrawals)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}
