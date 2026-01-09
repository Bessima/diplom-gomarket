package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/Bessima/diplom-gomarket/internal/config/db"
	"github.com/Bessima/diplom-gomarket/internal/customerror"
	"github.com/Bessima/diplom-gomarket/internal/models"
	"github.com/Bessima/diplom-gomarket/internal/retry"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type WithdrawRepository struct {
	db *db.DB
}

type WithdrawStorageRepositoryI interface {
	Create(userID int, orderID int64, sum int) error
	GetListByUserID(id int) ([]models.Withdrawal, error)
}

func NewWithdrawRepository(dbObj *db.DB) *WithdrawRepository {
	return &WithdrawRepository{db: dbObj}
}

func (repository *WithdrawRepository) Create(userID int, orderID int64, sum int) error {
	query := `INSERT INTO withdrawals (user_id,order_id, sum) VALUES ($1, $2, $3)`

	return retry.DoRetry(context.Background(), func() error {
		row, err := repository.db.Pool.Exec(context.Background(), query, userID, orderID, sum)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if pgErr.Code == pgerrcode.UniqueViolation {
					errWithMessage := fmt.Sprintf("withdraw with orderID %v already exists", orderID)
					return customerror.NewUniqueViolationError(errWithMessage)
				}
			}
			return customerror.NewCommonPGError(err.Error())
		}
		if row.RowsAffected() == 0 {
			err = fmt.Errorf("withdraw was not installed for orderID %v", orderID)
			return customerror.NewCommonPGError(err.Error())
		}
		return nil
	})
}

func (repository *WithdrawRepository) GetListByUserID(userID int) ([]models.Withdrawal, error) {
	query := `SELECT order_id,user_id,sum,processed_at FROM withdrawals WHERE user_id = $1`
	return retry.DoRetryWithResult(context.Background(), func() ([]models.Withdrawal, error) {
		rows, err := repository.db.Pool.Query(
			context.Background(),
			query,
			userID,
		)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		withdrawals := []models.Withdrawal{}
		for rows.Next() {
			var withdrawal models.Withdrawal
			var sumInKopecks int32
			err = rows.Scan(&withdrawal.OrderID, &withdrawal.UserID, &sumInKopecks, &withdrawal.ProcessedAt)

			if err != nil {
				return nil, err
			}

			withdrawal.SetSumInFloat(sumInKopecks)
			withdrawals = append(withdrawals, withdrawal)
		}

		err = rows.Err()
		if err != nil {
			return nil, err
		}

		return withdrawals, err
	})
}
