package repository

import (
	"errors"
	"testing"
	"time"

	"github.com/Bessima/diplom-gomarket/internal/customerror"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithdrawRepository_Create_Success(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	dbObj := NewTestDB(mock)
	repo := NewWithdrawRepository(dbObj)

	userID := 1
	orderID := int64(12345)
	sum := 10000

	mock.ExpectExec("INSERT INTO withdrawals").
		WithArgs(userID, orderID, sum).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	// Act
	err = repo.Create(userID, orderID, sum)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWithdrawRepository_Create_UniqueViolation(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	dbObj := NewTestDB(mock)
	repo := NewWithdrawRepository(dbObj)

	userID := 1
	orderID := int64(12345)
	sum := 10000

	pgErr := &pgconn.PgError{
		Code:    pgerrcode.UniqueViolation,
		Message: "duplicate key value",
	}

	mock.ExpectExec("INSERT INTO withdrawals").
		WithArgs(userID, orderID, sum).
		WillReturnError(pgErr)

	// Act
	err = repo.Create(userID, orderID, sum)

	// Assert
	assert.Error(t, err)
	assert.IsType(t, &customerror.UniqueViolationError{}, err)
	assert.Contains(t, err.Error(), "already exists")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWithdrawRepository_Create_NoRowsAffected(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	dbObj := NewTestDB(mock)
	repo := NewWithdrawRepository(dbObj)

	userID := 1
	orderID := int64(12345)
	sum := 10000

	mock.ExpectExec("INSERT INTO withdrawals").
		WithArgs(userID, orderID, sum).
		WillReturnResult(pgxmock.NewResult("INSERT", 0))

	// Act
	err = repo.Create(userID, orderID, sum)

	// Assert
	assert.Error(t, err)
	assert.IsType(t, &customerror.CommonPGError{}, err)
	assert.Contains(t, err.Error(), "was not installed")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWithdrawRepository_Create_DatabaseError(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	dbObj := NewTestDB(mock)
	repo := NewWithdrawRepository(dbObj)

	userID := 1
	orderID := int64(12345)
	sum := 10000

	pgErr := &pgconn.PgError{
		Code:    pgerrcode.ConnectionException,
		Message: "connection error",
	}

	mock.ExpectExec("INSERT INTO withdrawals").
		WithArgs(userID, orderID, sum).
		WillReturnError(pgErr)

	// Act
	err = repo.Create(userID, orderID, sum)

	// Assert
	assert.Error(t, err)
	assert.IsType(t, &customerror.CommonPGError{}, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWithdrawRepository_GetListByUserID_Success(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	dbObj := NewTestDB(mock)
	repo := NewWithdrawRepository(dbObj)

	userID := 1
	now := time.Now()

	rows := pgxmock.NewRows([]string{"order_id", "user_id", "sum", "processed_at"}).
		AddRow("12345", userID, int32(10000), now).
		AddRow("67890", userID, int32(25050), now.Add(-time.Hour))

	mock.ExpectQuery("SELECT order_id,user_id,sum,processed_at FROM withdrawals WHERE user_id").
		WithArgs(userID).
		WillReturnRows(rows)

	// Act
	withdrawals, err := repo.GetListByUserID(userID)

	// Assert
	assert.NoError(t, err)
	require.Len(t, withdrawals, 2)

	assert.Equal(t, "12345", withdrawals[0].OrderID)
	assert.Equal(t, userID, withdrawals[0].UserID)
	assert.Equal(t, float32(100.00), withdrawals[0].Sum)
	assert.Equal(t, now.Unix(), withdrawals[0].ProcessedAt.Unix())

	assert.Equal(t, "67890", withdrawals[1].OrderID)
	assert.Equal(t, userID, withdrawals[1].UserID)
	assert.Equal(t, float32(250.50), withdrawals[1].Sum)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWithdrawRepository_GetListByUserID_Empty(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	dbObj := NewTestDB(mock)
	repo := NewWithdrawRepository(dbObj)

	userID := 999

	rows := pgxmock.NewRows([]string{"order_id", "user_id", "sum", "processed_at"})

	mock.ExpectQuery("SELECT order_id,user_id,sum,processed_at FROM withdrawals WHERE user_id").
		WithArgs(userID).
		WillReturnRows(rows)

	// Act
	withdrawals, err := repo.GetListByUserID(userID)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, withdrawals)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWithdrawRepository_GetListByUserID_QueryError(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	dbObj := NewTestDB(mock)
	repo := NewWithdrawRepository(dbObj)

	userID := 1
	expectedError := errors.New("query error")

	mock.ExpectQuery("SELECT order_id,user_id,sum,processed_at FROM withdrawals WHERE user_id").
		WithArgs(userID).
		WillReturnError(expectedError)

	// Act
	withdrawals, err := repo.GetListByUserID(userID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, withdrawals)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWithdrawRepository_GetListByUserID_ScanError(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	dbObj := NewTestDB(mock)
	repo := NewWithdrawRepository(dbObj)

	userID := 1
	now := time.Now()

	rows := pgxmock.NewRows([]string{"order_id", "user_id", "sum", "processed_at"}).
		AddRow("invalid", userID, int32(10000), now).
		RowError(0, errors.New("scan error"))

	mock.ExpectQuery("SELECT order_id,user_id,sum,processed_at FROM withdrawals WHERE user_id").
		WithArgs(userID).
		WillReturnRows(rows)

	// Act
	withdrawals, err := repo.GetListByUserID(userID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, withdrawals)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWithdrawRepository_Create_DifferentAmounts(t *testing.T) {
	testCases := []struct {
		name    string
		userID  int
		orderID int64
		sum     int
	}{
		{
			name:    "small sum",
			userID:  1,
			orderID: 12345,
			sum:     100,
		},
		{
			name:    "large sum",
			userID:  2,
			orderID: 67890,
			sum:     100000,
		},
		{
			name:    "medium sum",
			userID:  3,
			orderID: 11111,
			sum:     50000,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			dbObj := NewTestDB(mock)
			repo := NewWithdrawRepository(dbObj)

			mock.ExpectExec("INSERT INTO withdrawals").
				WithArgs(tc.userID, tc.orderID, tc.sum).
				WillReturnResult(pgxmock.NewResult("INSERT", 1))

			// Act
			err = repo.Create(tc.userID, tc.orderID, tc.sum)

			// Assert
			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestWithdrawRepository_GetListByUserID_MultipleUsers(t *testing.T) {
	testCases := []struct {
		name        string
		userID      int
		withdrawals []struct {
			orderID string
			sum     int32
		}
	}{
		{
			name:   "user with one withdrawal",
			userID: 1,
			withdrawals: []struct {
				orderID string
				sum     int32
			}{
				{orderID: "12345", sum: 10000},
			},
		},
		{
			name:   "user with multiple withdrawals",
			userID: 2,
			withdrawals: []struct {
				orderID string
				sum     int32
			}{
				{orderID: "11111", sum: 5000},
				{orderID: "22222", sum: 15000},
				{orderID: "33333", sum: 25000},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			dbObj := NewTestDB(mock)
			repo := NewWithdrawRepository(dbObj)

			rows := pgxmock.NewRows([]string{"order_id", "user_id", "sum", "processed_at"})
			for _, w := range tc.withdrawals {
				rows.AddRow(w.orderID, tc.userID, w.sum, time.Now())
			}

			mock.ExpectQuery("SELECT order_id,user_id,sum,processed_at FROM withdrawals WHERE user_id").
				WithArgs(tc.userID).
				WillReturnRows(rows)

			// Act
			withdrawals, err := repo.GetListByUserID(tc.userID)

			// Assert
			assert.NoError(t, err)
			assert.Len(t, withdrawals, len(tc.withdrawals))
			for i, w := range tc.withdrawals {
				assert.Equal(t, w.orderID, withdrawals[i].OrderID)
				assert.Equal(t, tc.userID, withdrawals[i].UserID)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
