package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/Bessima/diplom-gomarket/internal/config/db"
	"github.com/Bessima/diplom-gomarket/internal/models"
	"github.com/Bessima/diplom-gomarket/internal/retry"
	"github.com/jackc/pgx/v5"
)

type BalanceRepository struct {
	db *db.DB
}

func NewBalanceRepository(dbObj *db.DB) *BalanceRepository {
	return &BalanceRepository{db: dbObj}
}

func (repository *BalanceRepository) GetBalanceUserID(userID int) (models.Balance, error) {
	query := `SELECT current, withdrawals FROM balance WHERE user_id = $1`
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
			if errors.Is(err, pgx.ErrNoRows) {
				// Считаем, что запрашиваемый пользователь есть, но пока не совершил покупок.
				// ? Возможно, при создании пользователя сразу стоит создавать таблицу баланса
				return balance, nil
			}
			return balance, err
		}

		balance.SetCurrent(Sum)
		balance.SetWithdrawn(Withdrawing)

		return balance, err
	})
}

func (repository *BalanceRepository) SetWithdrawForUserID(userID int, withdraw int) error {
	query := `UPDATE balance SET current = (balance.current - $1), withdrawals = (balance.withdrawals + $1) WHERE user_id=$2`
	return retry.DoRetry(context.Background(), func() error {
		row, err := repository.db.Pool.Exec(
			context.Background(),
			query,
			withdraw,
			userID,
		)
		if err != nil {
			return err
		}

		if row.RowsAffected() == 0 {
			errNoRow := fmt.Errorf("not update balance for user %d", userID)
			return errNoRow
		}

		return nil
	})
}

func (repository *BalanceRepository) SetAccrual(tx pgx.Tx, orderID string, userID int, accrual int32) error {
	queryBalance := `INSERT INTO balance (user_id, current) VALUES ($1,$2) ON CONFLICT (user_id) DO UPDATE SET current = balance.current + EXCLUDED.current`

	rowBalance, err := tx.Exec(
		context.Background(),
		queryBalance,
		userID,
		accrual,
	)
	if err != nil {
		return err
	}

	if rowBalance.RowsAffected() == 0 {
		err = fmt.Errorf("order with id %v was not installed accrual value", orderID)
		return err
	}
	return err
}
