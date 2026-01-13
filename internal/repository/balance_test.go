package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBalanceRepository_GetBalanceUserID_Success(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	dbObj := NewTestDB(mock)
	repo := NewBalanceRepository(dbObj)

	userID := 1
	currentSum := int32(50000)     // 500.00 рублей
	withdrawingSum := int32(10050) // 100.50 рублей

	rows := pgxmock.NewRows([]string{"current", "withdrawals"}).
		AddRow(currentSum, withdrawingSum)

	mock.ExpectQuery("SELECT current, withdrawals FROM balance WHERE user_id").
		WithArgs(userID).
		WillReturnRows(rows)

	// Act
	balance, err := repo.GetBalanceUserID(userID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, userID, balance.UserID)
	assert.Equal(t, float32(500.00), balance.Current)
	assert.Equal(t, float32(100.50), balance.Withdrawn)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBalanceRepository_GetBalanceUserID_NoRows(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	dbObj := NewTestDB(mock)
	repo := NewBalanceRepository(dbObj)

	userID := 999

	rows := pgxmock.NewRows([]string{"current", "withdrawals"})
	mock.ExpectQuery("SELECT current, withdrawals FROM balance WHERE user_id").
		WithArgs(userID).
		WillReturnRows(rows)

	// Act
	balance, err := repo.GetBalanceUserID(userID)

	// Assert
	assert.NoError(t, err) // Метод возвращает пустой баланс без ошибки
	assert.Equal(t, userID, balance.UserID)
	assert.Equal(t, float32(0), balance.Current)
	assert.Equal(t, float32(0), balance.Withdrawn)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBalanceRepository_GetBalanceUserID_DatabaseError(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	dbObj := NewTestDB(mock)
	repo := NewBalanceRepository(dbObj)

	userID := 1
	expectedError := errors.New("database connection error")

	mock.ExpectQuery("SELECT current, withdrawals FROM balance WHERE user_id").
		WithArgs(userID).
		WillReturnError(expectedError)

	// Act
	balance, err := repo.GetBalanceUserID(userID)

	// Assert
	assert.Error(t, err)
	// При ошибке возвращается zero value Balance
	assert.Equal(t, 0, balance.UserID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBalanceRepository_GetBalanceUserID_ZeroBalance(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	dbObj := NewTestDB(mock)
	repo := NewBalanceRepository(dbObj)

	userID := 1
	currentSum := int32(0)
	withdrawingSum := int32(0)

	rows := pgxmock.NewRows([]string{"current", "withdrawals"}).
		AddRow(currentSum, withdrawingSum)

	mock.ExpectQuery("SELECT current, withdrawals FROM balance WHERE user_id").
		WithArgs(userID).
		WillReturnRows(rows)

	// Act
	balance, err := repo.GetBalanceUserID(userID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, userID, balance.UserID)
	assert.Equal(t, float32(0), balance.Current)
	assert.Equal(t, float32(0), balance.Withdrawn)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBalanceRepository_SetWithdrawForUserID_Success(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	dbObj := NewTestDB(mock)
	repo := NewBalanceRepository(dbObj)

	userID := 1
	withdraw := 10000 // 100.00 рублей

	mock.ExpectExec("UPDATE balance SET current").
		WithArgs(withdraw, userID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	// Act
	err = repo.SetWithdrawForUserID(userID, withdraw)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBalanceRepository_SetWithdrawForUserID_NoRowsAffected(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	dbObj := NewTestDB(mock)
	repo := NewBalanceRepository(dbObj)

	userID := 999
	withdraw := 10000

	mock.ExpectExec("UPDATE balance SET current").
		WithArgs(withdraw, userID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	// Act
	err = repo.SetWithdrawForUserID(userID, withdraw)

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

	dbObj := NewTestDB(mock)
	repo := NewBalanceRepository(dbObj)

	userID := 1
	withdraw := 10000
	expectedError := errors.New("database error")

	mock.ExpectExec("UPDATE balance SET current").
		WithArgs(withdraw, userID).
		WillReturnError(expectedError)

	// Act
	err = repo.SetWithdrawForUserID(userID, withdraw)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBalanceRepository_SetAccrual_Success(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	dbObj := NewTestDB(mock)
	repo := NewBalanceRepository(dbObj)

	orderID := "12345"
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
	err = repo.SetAccrual(tx, orderID, userID, accrual)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBalanceRepository_SetAccrual_NoRowsAffected(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	dbObj := NewTestDB(mock)
	repo := NewBalanceRepository(dbObj)

	orderID := "12345"
	userID := 1
	accrual := int32(50000)

	mock.ExpectBegin()

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	mock.ExpectExec("INSERT INTO balance").
		WithArgs(userID, accrual).
		WillReturnResult(pgxmock.NewResult("INSERT", 0))

	// Act
	err = repo.SetAccrual(tx, orderID, userID, accrual)

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

	dbObj := NewTestDB(mock)
	repo := NewBalanceRepository(dbObj)

	orderID := "12345"
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
	err = repo.SetAccrual(tx, orderID, userID, accrual)

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

			dbObj := NewTestDB(mock)
			repo := NewBalanceRepository(dbObj)

			mock.ExpectExec("UPDATE balance SET current").
				WithArgs(tc.withdraw, tc.userID).
				WillReturnResult(pgxmock.NewResult("UPDATE", 1))

			// Act
			err = repo.SetWithdrawForUserID(tc.userID, tc.withdraw)

			// Assert
			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
