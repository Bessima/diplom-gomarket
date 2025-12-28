package repository

import (
	"context"
	"fmt"
	"github.com/Bessima/diplom-gomarket/internal/config/db"
	"github.com/Bessima/diplom-gomarket/internal/models"
	"github.com/Bessima/diplom-gomarket/internal/retry"
)

type OrderRepository struct {
	db *db.DB
}

type OrderStorageRepositoryI interface {
	Create(userID, orderID int) error
	GetByID(id int) (*models.Order, error)
	GetListByUserID(userID int) ([]models.Order, error)
	GetBalanceUserID(userID int) (models.Balance, error)
}

func NewOrderRepository(dbObj *db.DB) *OrderRepository {
	return &OrderRepository{db: dbObj}
}

func (repository *OrderRepository) Create(userID, orderID int) error {
	query := `INSERT INTO orders (id, user_id, status) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING`

	return retry.DoRetry(context.Background(), func() error {
		row, err := repository.db.Pool.Exec(context.Background(), query, orderID, userID, models.NewStatus)
		if err != nil {
			return err
		}
		if row.RowsAffected() == 0 {
			err = fmt.Errorf("order with id %v already exists", orderID)
		}
		return err
	})
}

func (repository *OrderRepository) GetByID(id int) (*models.Order, error) {
	query := `SELECT id,user_id,accrual,status FROM orders WHERE id = $1`
	return retry.DoRetryWithResult(context.Background(), func() (*models.Order, error) {
		row := repository.db.Pool.QueryRow(
			context.Background(),
			query,
			id,
		)

		elem := models.Order{}
		err := row.Scan(&elem.ID, &elem.UserID, &elem.Accrual, &elem.Status)
		return &elem, err
	})
}

func (repository *OrderRepository) GetListByUserID(userID int) ([]models.Order, error) {
	query := `SELECT id,user_id,accrual,status FROM orders WHERE user_id = $1`
	return retry.DoRetryWithResult(context.Background(), func() ([]models.Order, error) {
		rows, err := repository.db.Pool.Query(
			context.Background(),
			query,
			userID,
		)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		orders := []models.Order{}
		for rows.Next() {
			var order models.Order
			var accrualInKopecks *int32
			err = rows.Scan(&order.ID, &order.UserID, &accrualInKopecks, &order.Status)

			if err != nil {
				return nil, err
			}

			if accrualInKopecks != nil {
				order.SetAccrualAsFloat(*accrualInKopecks)
			}

			orders = append(orders, order)
		}

		err = rows.Err()
		if err != nil {
			return nil, err
		}

		return orders, err
	})
}

func (repository *OrderRepository) GetBalanceUserID(userID int) (models.Balance, error) {
	query := `SELECT user_id, current, withdrawals FROM balance WHERE user_id = $1`
	return retry.DoRetryWithResult(context.Background(), func() (models.Balance, error) {
		row := repository.db.Pool.QueryRow(
			context.Background(),
			query,
			userID,
		)
		balance := models.NewBalance(userID)

		var Sum int32
		var Withdrawing int32
		err := row.Scan(&Sum, &Withdrawing)
		if err != nil {
			return balance, err
		}

		balance.SetCurrent(Sum)
		balance.SetWithdrawing(Withdrawing)

		return balance, err
	})
}

func (repository *OrderRepository) SetListForProcessing(ch chan models.Order) error {
	query := `SELECT id,user_id,accrual,status FROM orders WHERE status IN ($1, $2)`
	return retry.DoRetry(context.Background(), func() error {

		rows, err := repository.db.Pool.Query(
			context.Background(),
			query,
			models.NewStatus,
			models.ProcessingStatus,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		var order models.Order

		for rows.Next() {
			err = rows.Scan(&order.ID, &order.UserID, &order.Accrual, &order.Status)

			if err != nil {
				return err
			}
			ch <- order
		}

		return rows.Err()
	})
}

func (repository *OrderRepository) UpdateStatus(orderID int, newStatus models.OrderStatus) error {
	query := `UPDATE orders SET status = $1 WHERE id = $2`
	return retry.DoRetry(context.Background(), func() error {

		row, err := repository.db.Pool.Exec(
			context.Background(),
			query,
			newStatus,
			orderID,
		)
		if row.RowsAffected() == 0 {
			err = fmt.Errorf("order with id %v already exists", orderID)
		}
		return err

	})
}

func (repository *OrderRepository) SetAccrual(orderID, userID int, accrual int32) error {
	ctx := context.Background()

	queryOrder := `UPDATE orders SET accrual = $1, status = $2 WHERE id = $3`
	queryBalance := `INSERT INTO balance (user_id, current) VALUES ($1,$2) ON CONFLICT (user_id) DO UPDATE SET current = balance.current + EXCLUDED.current`

	return retry.DoRetry(context.Background(), func() error {
		tx, err := repository.db.Pool.Begin(ctx)
		if err != nil {
			return err
		}
		defer func() {
			if err != nil {
				tx.Rollback(ctx)
			}
		}()

		row, err := tx.Exec(
			context.Background(),
			queryOrder,
			accrual,
			models.ProcessedStatus,
			orderID,
		)
		if row.RowsAffected() == 0 {
			err = fmt.Errorf("order with id %v was not installed accrual value", orderID)
			return err
		}

		rowBalance, err := tx.Exec(
			context.Background(),
			queryBalance,
			userID,
			accrual,
		)
		if rowBalance.RowsAffected() == 0 {
			err = fmt.Errorf("order with id %v was not installed accrual value", orderID)
			return err
		}
		return tx.Commit(ctx)
	})
}
