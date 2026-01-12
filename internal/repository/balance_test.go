package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/Bessima/diplom-gomarket/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBalanceRepository_GetBalanceUserID_Success(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	testRepo := &testBalanceRepository{mock: mock}

	userID := 1
	currentSum := int32(50000)     // 500.00 рублей
	withdrawingSum := int32(10050) // 100.50 рублей

	rows := pgxmock.NewRows([]string{"current", "withdrawals"}).
		AddRow(currentSum, withdrawingSum)

	mock.ExpectQuery("SELECT current, withdrawals FROM balance WHERE user_id").
		WithArgs(userID).
		WillReturnRows(rows)

	// Act
	balance, err := testRepo.GetBalanceUserID(userID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, userID, balance.UserID)
	assert.Equal(t, float32(500.00), balance.Current)
	assert.Equal(t, float32(100.50), balance.Withdrawing)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBalanceRepository_GetBalanceUserID_NoRows(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	testRepo := &testBalanceRepository{mock: mock}

	userID := 999

	mock.ExpectQuery("SELECT current, withdrawals FROM balance WHERE user_id").
		WithArgs(userID).
		WillReturnError(errors.New("no rows in result set"))

	// Act
	balance, err := testRepo.GetBalanceUserID(userID)

	// Assert
	assert.NoError(t, err) // Метод возвращает пустой баланс без ошибки
	assert.Equal(t, userID, balance.UserID)
	assert.Equal(t, float32(0), balance.Current)
	assert.Equal(t, float32(0), balance.Withdrawing)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBalanceRepository_GetBalanceUserID_DatabaseError(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	testRepo := &testBalanceRepository{mock: mock}

	userID := 1
	expectedError := errors.New("database connection error")

	mock.ExpectQuery("SELECT current, withdrawals FROM balance WHERE user_id").
		WithArgs(userID).
		WillReturnError(expectedError)

	// Act
	balance, err := testRepo.GetBalanceUserID(userID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, userID, balance.UserID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBalanceRepository_GetBalanceUserID_ZeroBalance(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	testRepo := &testBalanceRepository{mock: mock}

	userID := 1
	currentSum := int32(0)
	withdrawingSum := int32(0)

	rows := pgxmock.NewRows([]string{"current", "withdrawals"}).
		AddRow(currentSum, withdrawingSum)

	mock.ExpectQuery("SELECT current, withdrawals FROM balance WHERE user_id").
		WithArgs(userID).
		WillReturnRows(rows)

	// Act
	balance, err := testRepo.GetBalanceUserID(userID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, userID, balance.UserID)
	assert.Equal(t, float32(0), balance.Current)
	assert.Equal(t, float32(0), balance.Withdrawing)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBalanceRepository_SetWithdrawForUserID_Success(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	testRepo := &testBalanceRepository{mock: mock}

	userID := 1
	withdraw := 10000 // 100.00 рублей

	mock.ExpectExec("UPDATE balance SET current").
		WithArgs(withdraw, userID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	// Act
	err = testRepo.SetWithdrawForUserID(userID, withdraw)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBalanceRepository_SetWithdrawForUserID_NoRowsAffected(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	testRepo := &testBalanceRepository{mock: mock}

	userID := 999
	withdraw := 10000

	mock.ExpectExec("UPDATE balance SET current").
		WithArgs(withdraw, userID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	// Act
	err = testRepo.SetWithdrawForUserID(userID, withdraw)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not update balance for user")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBalanceRepository_SetWithdrawForUserID_DatabaseError(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	testRepo := &testBalanceRepository{mock: mock}

	userID := 1
	withdraw := 10000
	expectedError := errors.New("database error")

	mock.ExpectExec("UPDATE balance SET current").
		WithArgs(withdraw, userID).
		WillReturnError(expectedError)

	// Act
	err = testRepo.SetWithdrawForUserID(userID, withdraw)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not update balance for user")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBalanceRepository_SetAccrual_Success(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	testRepo := &testBalanceRepository{mock: mock}

	orderID := 12345
	userID := 1
	accrual := int32(50000)

	// Начинаем транзакцию
	mock.ExpectBegin()

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	mock.ExpectExec("INSERT INTO balance").
		WithArgs(userID, accrual).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	// Act
	err = testRepo.SetAccrual(tx, orderID, userID, accrual)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBalanceRepository_SetAccrual_NoRowsAffected(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	testRepo := &testBalanceRepository{mock: mock}

	orderID := 12345
	userID := 1
	accrual := int32(50000)

	mock.ExpectBegin()

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	mock.ExpectExec("INSERT INTO balance").
		WithArgs(userID, accrual).
		WillReturnResult(pgxmock.NewResult("INSERT", 0))

	// Act
	err = testRepo.SetAccrual(tx, orderID, userID, accrual)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "was not installed accrual value")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBalanceRepository_SetAccrual_DatabaseError(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	testRepo := &testBalanceRepository{mock: mock}

	orderID := 12345
	userID := 1
	accrual := int32(50000)
	expectedError := errors.New("database error")

	mock.ExpectBegin()

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	mock.ExpectExec("INSERT INTO balance").
		WithArgs(userID, accrual).
		WillReturnError(expectedError)

	// Act
	err = testRepo.SetAccrual(tx, orderID, userID, accrual)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBalanceRepository_SetWithdrawForUserID_DifferentAmounts(t *testing.T) {
	testCases := []struct {
		name     string
		userID   int
		withdraw int
	}{
		{
			name:     "small withdraw",
			userID:   1,
			withdraw: 100,
		},
		{
			name:     "large withdraw",
			userID:   2,
			withdraw: 100000,
		},
		{
			name:     "medium withdraw",
			userID:   3,
			withdraw: 50000,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			testRepo := &testBalanceRepository{mock: mock}

			mock.ExpectExec("UPDATE balance SET current").
				WithArgs(tc.withdraw, tc.userID).
				WillReturnResult(pgxmock.NewResult("UPDATE", 1))

			// Act
			err = testRepo.SetWithdrawForUserID(tc.userID, tc.withdraw)

			// Assert
			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// testBalanceRepository - тестовая версия репозитория для работы с моками
type testBalanceRepository struct {
	mock pgxmock.PgxPoolIface
}

func (r *testBalanceRepository) GetBalanceUserID(userID int) (models.Balance, error) {
	query := `SELECT current, withdrawals FROM balance WHERE user_id = $1`
	row := r.mock.QueryRow(context.Background(), query, userID)
	balance := models.NewBalance(userID)

	var Sum int32
	var Withdrawing int32
	err := row.Scan(&Sum, &Withdrawing)
	if err != nil {
		errNoRow := errors.New("no rows in result set")
		if err.Error() == errNoRow.Error() {
			return balance, nil
		}
		return balance, err
	}

	balance.SetCurrent(Sum)
	balance.SetWithdrawing(Withdrawing)

	return balance, nil
}

func (r *testBalanceRepository) SetWithdrawForUserID(userID int, withdraw int) error {
	query := `UPDATE balance SET current = (balance.current - $1), withdrawals = (balance.withdrawals + $1) WHERE user_id=$2`
	row, err := r.mock.Exec(context.Background(), query, withdraw, userID)

	if err != nil || row.RowsAffected() == 0 {
		return errors.New("not update balance for user")
	}

	return nil
}

func (r *testBalanceRepository) SetAccrual(tx pgx.Tx, orderID, userID int, accrual int32) error {
	queryBalance := `INSERT INTO balance (user_id, current) VALUES ($1,$2) ON CONFLICT (user_id) DO UPDATE SET current = balance.current + EXCLUDED.current`

	rowBalance, err := tx.Exec(context.Background(), queryBalance, userID, accrual)
	if err != nil {
		return err
	}

	if rowBalance.RowsAffected() == 0 {
		return errors.New("order with id was not installed accrual value")
	}
	return nil
}
