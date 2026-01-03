package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Bessima/diplom-gomarket/internal/middlewares/logger"
	"github.com/Bessima/diplom-gomarket/internal/models"
	"github.com/Bessima/diplom-gomarket/internal/repository"
	"io"
	"net/http"
	"strconv"
)

type OrdersHandler struct {
	OrderStorage        repository.OrderStorageRepositoryI
	BalanceStorage      *repository.BalanceRepository
	ordersForProcessing chan models.Order
}

func NewOrderHandler(storage repository.OrderStorageRepositoryI, balanceRepository *repository.BalanceRepository, ordersForProcessing chan models.Order) *OrdersHandler {
	return &OrdersHandler{
		OrderStorage:        storage,
		BalanceStorage:      balanceRepository,
		ordersForProcessing: ordersForProcessing,
	}
}

func (h *OrdersHandler) Add(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "can't read body", http.StatusBadRequest)
		logger.Log.Error(err.Error())
		return
	}

	bodyString := string(bodyBytes)
	if !CheckLuhn(bodyString) {
		http.Error(w, "invalid order number", http.StatusUnprocessableEntity)
		logger.Log.Error(fmt.Sprintf("invalid order number: %s", bodyString))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	orderID, err := strconv.Atoi(bodyString)
	if err != nil {
		http.Error(w, "invalid order number", http.StatusBadRequest)
		return
	}

	user := GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "user was not got", http.StatusBadRequest)
		logger.Log.Error("user was not got")
		return
	}
	order, _ := h.OrderStorage.GetByID(orderID)
	if order != nil {
		if order.UserID == user.ID {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("order already added"))
			return
		}

		http.Error(w, "order was already added other user", http.StatusConflict)
		logger.Log.Warn(fmt.Sprintf("order %d was already added other user", orderID))
		return
	}

	err = h.OrderStorage.Create(user.ID, orderID)

	if err != nil {
		http.Error(w, "order was not created", http.StatusInternalServerError)
		logger.Log.Warn(fmt.Sprintf("order was not created, error: %v", err))
		return
	}
	h.ordersForProcessing <- models.Order{ID: orderID, UserID: user.ID, Status: models.NewStatus}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Order added successfully!"))
}

func (h *OrdersHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "user was not got", http.StatusBadRequest)
		logger.Log.Error("user was not got")
		return
	}

	orders, err := h.OrderStorage.GetListByUserID(user.ID)
	if err != nil {
		http.Error(w, "orders were not found", http.StatusInternalServerError)
		logger.Log.Error(fmt.Sprintf("orders were not found, error: %v", err))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(orders)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)

}

func (h *OrdersHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "user was not got", http.StatusBadRequest)
		logger.Log.Error("user was not got")
		return
	}

	balance, err := h.BalanceStorage.GetBalanceUserID(user.ID)
	if err != nil {
		errNoRow := errors.New("no rows in result set")
		if err.Error() != errNoRow.Error() {
			http.Error(w, "Error to getting orders", http.StatusInternalServerError)
			logger.Log.Warn(fmt.Sprintf("Error to getting orders, error: %v", err))
			return
		}
	}

	w.Header().Add("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(balance)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}
