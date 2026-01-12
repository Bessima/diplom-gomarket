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

func TestOrderRepository_Create_Success(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	// Создаем репозиторий напрямую с моком, используя testOrderRepository
	testRepo := &testOrderRepository{mock: mock}

	userID := 1
	orderID := 12345

	// Ожидаем один запрос INSERT
	mock.ExpectExec("INSERT INTO orders").
		WithArgs(orderID, userID, models.NewStatus).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	// Act
	err = testRepo.Create(userID, orderID)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_Create_AlreadyExists(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	testRepo := &testOrderRepository{mock: mock}

	userID := 1
	orderID := 12345

	// Имитируем конфликт - 0 затронутых строк (ON CONFLICT DO NOTHING)
	mock.ExpectExec("INSERT INTO orders").
		WithArgs(orderID, userID, models.NewStatus).
		WillReturnResult(pgxmock.NewResult("INSERT", 0))

	// Act
	err = testRepo.Create(userID, orderID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_Create_DatabaseError(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	testRepo := &testOrderRepository{mock: mock}

	userID := 1
	orderID := 12345
	expectedError := errors.New("database connection error")

	mock.ExpectExec("INSERT INTO orders").
		WithArgs(orderID, userID, models.NewStatus).
		WillReturnError(expectedError)

	// Act
	err = testRepo.Create(userID, orderID)

	// Assert
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_GetByID_Success(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	testRepo := &testOrderRepository{mock: mock}

	orderID := 12345
	userID := 1
	accrualInt := int32(50050)
	status := models.ProcessedStatus

	// Используем указатель на int32 для accrual
	accrualPtr := &accrualInt

	rows := pgxmock.NewRows([]string{"id", "user_id", "accrual", "status"}).
		AddRow(orderID, userID, accrualPtr, status)

	mock.ExpectQuery("SELECT id,user_id,accrual,status FROM orders WHERE id").
		WithArgs(orderID).
		WillReturnRows(rows)

	// Act
	order, err := testRepo.GetByID(orderID)

	// Assert
	assert.NoError(t, err)
	require.NotNil(t, order)
	assert.Equal(t, orderID, order.ID)
	assert.Equal(t, userID, order.UserID)
	assert.NotNil(t, order.Accrual)
	assert.Equal(t, float32(500.50), *order.Accrual)
	assert.Equal(t, status, order.Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_GetByID_NotFound(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	testRepo := &testOrderRepository{mock: mock}

	orderID := 99999

	mock.ExpectQuery("SELECT id,user_id,accrual,status FROM orders WHERE id").
		WithArgs(orderID).
		WillReturnError(pgx.ErrNoRows)

	// Act
	order, err := testRepo.GetByID(orderID)

	// Assert
	assert.Error(t, err)
	require.NotNil(t, order)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_GetListByUserID_Success(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	testRepo := &testOrderRepository{mock: mock}

	userID := 1
	accrual1 := int32(10050)
	accrual2 := int32(20000)
	accrual1Ptr := &accrual1
	accrual2Ptr := &accrual2

	rows := pgxmock.NewRows([]string{"id", "user_id", "accrual", "status"}).
		AddRow(12345, userID, accrual1Ptr, models.ProcessedStatus).
		AddRow(67890, userID, accrual2Ptr, models.ProcessingStatus)

	mock.ExpectQuery("SELECT id,user_id,accrual,status FROM orders WHERE user_id").
		WithArgs(userID).
		WillReturnRows(rows)

	// Act
	orders, err := testRepo.GetListByUserID(userID)

	// Assert
	assert.NoError(t, err)
	require.Len(t, orders, 2)

	assert.Equal(t, 12345, orders[0].ID)
	assert.Equal(t, userID, orders[0].UserID)
	assert.NotNil(t, orders[0].Accrual)
	assert.Equal(t, float32(100.50), *orders[0].Accrual)
	assert.Equal(t, models.ProcessedStatus, orders[0].Status)

	assert.Equal(t, 67890, orders[1].ID)
	assert.Equal(t, userID, orders[1].UserID)
	assert.NotNil(t, orders[1].Accrual)
	assert.Equal(t, float32(200.00), *orders[1].Accrual)
	assert.Equal(t, models.ProcessingStatus, orders[1].Status)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_GetListByUserID_Empty(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	testRepo := &testOrderRepository{mock: mock}

	userID := 999

	rows := pgxmock.NewRows([]string{"id", "user_id", "accrual", "status"})

	mock.ExpectQuery("SELECT id,user_id,accrual,status FROM orders WHERE user_id").
		WithArgs(userID).
		WillReturnRows(rows)

	// Act
	orders, err := testRepo.GetListByUserID(userID)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, orders)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_GetListByUserID_WithNullAccrual(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	testRepo := &testOrderRepository{mock: mock}

	userID := 1

	rows := pgxmock.NewRows([]string{"id", "user_id", "accrual", "status"}).
		AddRow(12345, userID, nil, models.NewStatus)

	mock.ExpectQuery("SELECT id,user_id,accrual,status FROM orders WHERE user_id").
		WithArgs(userID).
		WillReturnRows(rows)

	// Act
	orders, err := testRepo.GetListByUserID(userID)

	// Assert
	assert.NoError(t, err)
	require.Len(t, orders, 1)
	assert.Equal(t, 12345, orders[0].ID)
	assert.Nil(t, orders[0].Accrual)
	assert.Equal(t, models.NewStatus, orders[0].Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_GetListByUserID_QueryError(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	testRepo := &testOrderRepository{mock: mock}

	userID := 1
	expectedError := errors.New("query error")

	mock.ExpectQuery("SELECT id,user_id,accrual,status FROM orders WHERE user_id").
		WithArgs(userID).
		WillReturnError(expectedError)

	// Act
	orders, err := testRepo.GetListByUserID(userID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, orders)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_SetListForProcessing_Success(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	testRepo := &testOrderRepository{mock: mock}

	rows := pgxmock.NewRows([]string{"id", "user_id", "accrual", "status"}).
		AddRow(12345, 1, nil, models.NewStatus).
		AddRow(67890, 2, nil, models.ProcessingStatus)

	mock.ExpectQuery("SELECT id,user_id,accrual,status FROM orders WHERE status IN").
		WithArgs(models.NewStatus, models.ProcessingStatus).
		WillReturnRows(rows)

	ch := make(chan models.Order, 2)
	defer close(ch)

	// Act
	err = testRepo.SetListForProcessing(ch)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 2, len(ch))

	order1 := <-ch
	assert.Equal(t, 12345, order1.ID)
	assert.Equal(t, models.NewStatus, order1.Status)

	order2 := <-ch
	assert.Equal(t, 67890, order2.ID)
	assert.Equal(t, models.ProcessingStatus, order2.Status)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_SetListForProcessing_QueryError(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	testRepo := &testOrderRepository{mock: mock}

	expectedError := errors.New("query error")

	mock.ExpectQuery("SELECT id,user_id,accrual,status FROM orders WHERE status IN").
		WithArgs(models.NewStatus, models.ProcessingStatus).
		WillReturnError(expectedError)

	ch := make(chan models.Order, 2)
	defer close(ch)

	// Act
	err = testRepo.SetListForProcessing(ch)

	// Assert
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_UpdateStatus_Success(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	testRepo := &testOrderRepository{mock: mock}

	orderID := 12345
	newStatus := models.ProcessedStatus

	mock.ExpectExec("UPDATE orders SET status").
		WithArgs(newStatus, orderID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	// Act
	err = testRepo.UpdateStatus(orderID, newStatus)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_UpdateStatus_OrderNotFound(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	testRepo := &testOrderRepository{mock: mock}

	orderID := 99999
	newStatus := models.ProcessedStatus

	mock.ExpectExec("UPDATE orders SET status").
		WithArgs(newStatus, orderID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	// Act
	err = testRepo.UpdateStatus(orderID, newStatus)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_UpdateStatus_DatabaseError(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	testRepo := &testOrderRepository{mock: mock}

	orderID := 12345
	newStatus := models.ProcessedStatus
	expectedError := errors.New("database error")

	mock.ExpectExec("UPDATE orders SET status").
		WithArgs(newStatus, orderID).
		WillReturnError(expectedError)

	// Act
	err = testRepo.UpdateStatus(orderID, newStatus)

	// Assert
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_SetAccrual_Success(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	testRepo := &testOrderRepository{mock: mock}

	orderID := 12345
	userID := 1
	accrual := int32(50000)

	// Ожидаем начало транзакции
	mock.ExpectBegin()

	// Ожидаем обновление заказа
	mock.ExpectExec("UPDATE orders SET accrual").
		WithArgs(accrual, models.ProcessedStatus, orderID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	// Ожидаем обновление баланса
	mock.ExpectExec("INSERT INTO balance").
		WithArgs(userID, accrual).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	// Ожидаем коммит транзакции
	mock.ExpectCommit()

	// Act
	err = testRepo.SetAccrual(orderID, userID, accrual)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_SetAccrual_OrderNotFound(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	testRepo := &testOrderRepository{mock: mock}

	orderID := 99999
	userID := 1
	accrual := int32(50000)

	mock.ExpectBegin()

	// Заказ не найден - 0 затронутых строк
	mock.ExpectExec("UPDATE orders SET accrual").
		WithArgs(accrual, models.ProcessedStatus, orderID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	mock.ExpectRollback()

	// Act
	err = testRepo.SetAccrual(orderID, userID, accrual)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "was not installed accrual value")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_SetAccrual_BalanceUpdateError(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	testRepo := &testOrderRepository{mock: mock}

	orderID := 12345
	userID := 1
	accrual := int32(50000)
	expectedError := errors.New("balance update error")

	mock.ExpectBegin()

	mock.ExpectExec("UPDATE orders SET accrual").
		WithArgs(accrual, models.ProcessedStatus, orderID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	mock.ExpectExec("INSERT INTO balance").
		WithArgs(userID, accrual).
		WillReturnError(expectedError)

	mock.ExpectRollback()

	// Act
	err = testRepo.SetAccrual(orderID, userID, accrual)

	// Assert
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_SetAccrual_TransactionBeginError(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	testRepo := &testOrderRepository{mock: mock}

	orderID := 12345
	userID := 1
	accrual := int32(50000)
	expectedError := errors.New("transaction begin error")

	mock.ExpectBegin().WillReturnError(expectedError)

	// Act
	err = testRepo.SetAccrual(orderID, userID, accrual)

	// Assert
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_Create_WithDifferentOrderIDs(t *testing.T) {
	testCases := []struct {
		name     string
		userID   int
		orderID  int
		affected int64
		wantErr  bool
	}{
		{
			name:     "success creation",
			userID:   1,
			orderID:  12345,
			affected: 1,
			wantErr:  false,
		},
		{
			name:     "order already exists",
			userID:   2,
			orderID:  67890,
			affected: 0,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			testRepo := &testOrderRepository{mock: mock}

			mock.ExpectExec("INSERT INTO orders").
				WithArgs(tc.orderID, tc.userID, models.NewStatus).
				WillReturnResult(pgxmock.NewResult("INSERT", tc.affected))

			// Act
			err = testRepo.Create(tc.userID, tc.orderID)

			// Assert
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestOrderRepository_UpdateStatus_AllStatuses(t *testing.T) {
	statuses := []models.OrderStatus{
		models.NewStatus,
		models.InvalidStatus,
		models.ProcessingStatus,
		models.RegisterAcSystemStatus,
		models.ProcessedStatus,
	}

	for _, status := range statuses {
		t.Run(string(status), func(t *testing.T) {
			// Arrange
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			testRepo := &testOrderRepository{mock: mock}

			orderID := 12345

			mock.ExpectExec("UPDATE orders SET status").
				WithArgs(status, orderID).
				WillReturnResult(pgxmock.NewResult("UPDATE", 1))

			// Act
			err = testRepo.UpdateStatus(orderID, status)

			// Assert
			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// testOrderRepository - тестовая версия репозитория для работы с моками
type testOrderRepository struct {
	mock pgxmock.PgxPoolIface
}

func (r *testOrderRepository) Create(userID, orderID int) error {
	query := `INSERT INTO orders (id, user_id, status) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING`
	row, err := r.mock.Exec(context.Background(), query, orderID, userID, models.NewStatus)
	if err != nil {
		return err
	}
	if row.RowsAffected() == 0 {
		return errors.New("order with id already exists")
	}
	return nil
}

func (r *testOrderRepository) GetByID(id int) (*models.Order, error) {
	query := `SELECT id,user_id,accrual,status FROM orders WHERE id = $1`
	row := r.mock.QueryRow(context.Background(), query, id)

	elem := models.Order{}
	var accrualInKopecks *int32
	err := row.Scan(&elem.ID, &elem.UserID, &accrualInKopecks, &elem.Status)
	if err != nil {
		return &elem, err
	}

	if accrualInKopecks != nil {
		elem.SetAccrualAsFloat(*accrualInKopecks)
	}

	return &elem, nil
}

func (r *testOrderRepository) GetListByUserID(userID int) ([]models.Order, error) {
	query := `SELECT id,user_id,accrual,status FROM orders WHERE user_id = $1`
	rows, err := r.mock.Query(context.Background(), query, userID)
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

	return orders, nil
}

func (r *testOrderRepository) SetListForProcessing(ch chan models.Order) error {
	query := `SELECT id,user_id,accrual,status FROM orders WHERE status IN ($1, $2)`
	rows, err := r.mock.Query(context.Background(), query, models.NewStatus, models.ProcessingStatus)
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
}

func (r *testOrderRepository) UpdateStatus(orderID int, newStatus models.OrderStatus) error {
	query := `UPDATE orders SET status = $1 WHERE id = $2`
	row, err := r.mock.Exec(context.Background(), query, newStatus, orderID)
	if row.RowsAffected() == 0 {
		return errors.New("order with id already exists")
	}
	return err
}

func (r *testOrderRepository) SetAccrual(orderID, userID int, accrual int32) error {
	ctx := context.Background()

	queryOrder := `UPDATE orders SET accrual = $1, status = $2 WHERE id = $3`

	tx, err := r.mock.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	row, err := tx.Exec(ctx, queryOrder, accrual, models.ProcessedStatus, orderID)
	if err == nil && row.RowsAffected() == 0 {
		err = errors.New("order with id was not installed accrual value")
		return err
	}
	if err != nil {
		return err
	}

	// Имитируем вызов балансового репозитория
	queryBalance := `INSERT INTO balance (user_id, current) VALUES ($1,$2) ON CONFLICT (user_id) DO UPDATE SET current = balance.current + EXCLUDED.current`
	_, err = tx.Exec(ctx, queryBalance, userID, accrual)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
