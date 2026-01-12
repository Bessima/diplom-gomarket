package service

import (
	"context"
	"fmt"
	"github.com/Bessima/diplom-gomarket/internal/clients/accrual"
	"github.com/Bessima/diplom-gomarket/internal/config/db"
	"github.com/Bessima/diplom-gomarket/internal/middlewares/logger"
	"github.com/Bessima/diplom-gomarket/internal/models"
	"github.com/Bessima/diplom-gomarket/internal/repository"
	"time"
)

type OrderRepositoryI interface {
	UpdateStatus(orderID int, newStatus models.OrderStatus) error
	SetAccrual(orderID, userID int, accrual int32) error
	SetListForProcessing(ch chan models.Order) error
}

type OrderService struct {
	repository    OrderRepositoryI
	accrualClient accrual.AccrualClientI
}

func NewOrderService(dbObj *db.DB, accrualAddress string) *OrderService {
	rep := repository.NewOrderRepository(dbObj)
	accrualClient := accrual.NewAccrualClient(accrualAddress)

	return &OrderService{repository: rep, accrualClient: accrualClient}
}

func (service OrderService) GetAccrualForOrder(ctx context.Context, ordersForProcessing chan models.Order) {

	for order := range ordersForProcessing {
		resp, err := service.accrualClient.Get(ctx, order.ID)
		if err != nil {
			logger.Log.Warn(err.Error())
			//Заказы всегда будут в канале, если цель не достигнута
			time.Sleep(10 * time.Second)
			ordersForProcessing <- order
			continue
		}
		switch newStatus := models.OrderStatus(resp.Status); newStatus {
		case models.InvalidStatus:
			logger.Log.Info(fmt.Sprintf("Order %d has invalid status", order.ID))
			err = service.repository.UpdateStatus(order.ID, newStatus)
			if err != nil {
				logger.Log.Warn(err.Error())
			}
			continue
		case models.ProcessedStatus:
			logger.Log.Info(fmt.Sprintf("Order %d has already processed status", order.ID))
			accrualInt := int32(resp.Accrual * 100)
			err = service.repository.SetAccrual(order.ID, order.UserID, accrualInt)
			if err != nil {
				logger.Log.Warn(fmt.Sprintf("Order %d was not saved in DB, %s", order.ID, err.Error()))
				ordersForProcessing <- order
				continue
			}

		case models.ProcessingStatus:
		case models.RegisterAcSystemStatus:
			if newStatus != order.Status && newStatus != models.RegisterAcSystemStatus {
				err = service.repository.UpdateStatus(order.ID, newStatus)
				if err != nil {
					logger.Log.Warn(err.Error())
				}
				order.Status = newStatus
			}
			ordersForProcessing <- order
		}
	}
}

func (service OrderService) AddNotProcessedOrders(ordersForProcessing chan models.Order) {
	err := service.repository.SetListForProcessing(ordersForProcessing)
	if err != nil {
		logger.Log.Warn(err.Error())
	}
}
