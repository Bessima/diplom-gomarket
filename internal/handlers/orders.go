package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/Bessima/diplom-gomarket/internal/handlers/schemas"
	"github.com/Bessima/diplom-gomarket/internal/middlewares/logger"
	"github.com/Bessima/diplom-gomarket/internal/repository"
	"io"
	"net/http"
	"strconv"
)

type OrdersHandler struct {
	OrderStorage repository.OrderStorageRepositoryI
}

func NewOrderHandler(storage repository.OrderStorageRepositoryI) *OrdersHandler {
	return &OrdersHandler{OrderStorage: storage}
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
	if !checkLuhn(bodyString) {
		http.Error(w, "invalid order number", http.StatusBadRequest)
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
	err = h.OrderStorage.Create(user.ID, orderID)
	if err != nil {
		http.Error(w, "order was not created", http.StatusInternalServerError)
		logger.Log.Error(fmt.Sprintf("order was not created, error: %v", err))
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]string{"message": "Order added successfully!"})
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
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

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(orders)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

func checkLuhn(number string) bool {
	sum := 0
	double := false
	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')

		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		double = !double
	}
	return sum%10 == 0
}

func (h *OrdersHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "user was not got", http.StatusBadRequest)
		logger.Log.Error("user was not got")
		return
	}

	balance, err := h.OrderStorage.GetBalanceUserID(user.ID)
	if err != nil {
		http.Error(w, "orders were not found", http.StatusInternalServerError)
		logger.Log.Error(fmt.Sprintf("orders were not found, error: %v", err))
		return
	}
	schema := schemas.BalanceResponse{
		Current:   balance,
		Withdrawn: 0,
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(schema)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}
