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
	Create(user_id, order_id int) error
	GetByID(id int) (*models.Order, error)
}

func NewOrderRepository(dbObj *db.DB) *OrderRepository {
	return &OrderRepository{db: dbObj}
}

func (repository *OrderRepository) Create(user_id, order_id int) error {
	query := `INSERT INTO orders (id, user_id, status) VALUES ($1, $2, $3)`

	return retry.DoRetry(context.Background(), func() error {
		row, err := repository.db.Pool.Exec(context.Background(), query, order_id, user_id, models.NewStatus)
		if err != nil {
			return err
		}
		if row.RowsAffected() == 0 {
			err = fmt.Errorf("Order with id %v already exists", order_id)
		}
		return err
	})
}

func (repository *OrderRepository) GetByID(id int) (*models.Order, error) {
	query := `SELECT (id,user_id,amount,accrual,status) FROM orders WHERE id = $1`
	return retry.DoRetryWithResult(context.Background(), func() (*models.Order, error) {
		row := repository.db.Pool.QueryRow(
			context.Background(),
			query,
			id,
		)

		elem := models.Order{}
		err := row.Scan(&elem.ID, &elem.UserID, &elem.Amount, &elem.Accrual, &elem.Status)
		return &elem, err
	})
}
