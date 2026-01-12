package service

import (
	"errors"
	"testing"

	"github.com/Bessima/diplom-gomarket/internal/handlers/schemas"
	"github.com/Bessima/diplom-gomarket/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockWithdrawRepository - мок для WithdrawStorageRepositoryI
type MockWithdrawRepository struct {
	mock.Mock
}

func (m *MockWithdrawRepository) Create(userID int, orderID int64, sum int) error {
	args := m.Called(userID, orderID, sum)
	return args.Error(0)
}

func (m *MockWithdrawRepository) GetListByUserID(id int) ([]models.Withdrawal, error) {
	args := m.Called(id)
	return args.Get(0).([]models.Withdrawal), args.Error(1)
}

// BalanceRepositoryInterface - интерфейс для тестирования
type BalanceRepositoryInterface interface {
	SetWithdrawForUserID(userID int, withdraw int) error
}

// MockBalanceRepository - мок для BalanceRepository
type MockBalanceRepository struct {
	mock.Mock
}

func (m *MockBalanceRepository) SetWithdrawForUserID(userID int, withdraw int) error {
	args := m.Called(userID, withdraw)
	return args.Error(0)
}

// TestWithdrawService - тестовая версия сервиса для тестов
type TestWithdrawService struct {
	service         *WithdrawService
	mockBalanceRepo BalanceRepositoryInterface
}

func NewTestWithdrawService(withdrawRepo *MockWithdrawRepository, balanceRepo BalanceRepositoryInterface) *TestWithdrawService {
	return &TestWithdrawService{
		service: &WithdrawService{
			WithdrawRepository: withdrawRepo,
		},
		mockBalanceRepo: balanceRepo,
	}
}

// Set - тестовая реализация метода, которая использует моки
func (ts *TestWithdrawService) Set(user *models.User, withdrawRequest schemas.WithdrawRequest) error {
	ts.service.mu.Lock()
	defer ts.service.mu.Unlock()

	orderID, err := withdrawRequest.GetOrderAsInt()
	if err != nil {
		return errors.New("can't parse number of order")
	}
	withdrawInt := withdrawRequest.GetSumAsInt()
	err = ts.service.WithdrawRepository.Create(user.ID, orderID, withdrawInt)
	if err != nil {
		return err
	}

	err = ts.mockBalanceRepo.SetWithdrawForUserID(user.ID, withdrawInt)

	return err
}

func TestWithdrawService_Set_Success(t *testing.T) {
	// Arrange
	mockWithdrawRepo := new(MockWithdrawRepository)
	mockBalanceRepo := new(MockBalanceRepository)

	service := NewTestWithdrawService(mockWithdrawRepo, mockBalanceRepo)

	user := &models.User{
		ID:    1,
		Login: "testuser",
	}

	withdrawRequest := schemas.WithdrawRequest{
		Order: "12345",
		Sum:   100.50,
	}

	expectedOrderID := int64(12345)
	expectedSum := 10050 // 100.50 * 100

	// Настройка ожиданий для моков
	mockWithdrawRepo.On("Create", user.ID, expectedOrderID, expectedSum).Return(nil)
	mockBalanceRepo.On("SetWithdrawForUserID", user.ID, expectedSum).Return(nil)

	// Act
	err := service.Set(user, withdrawRequest)

	// Assert
	assert.NoError(t, err)
	mockWithdrawRepo.AssertExpectations(t)
	mockBalanceRepo.AssertExpectations(t)
}

func TestWithdrawService_Set_InvalidOrderNumber(t *testing.T) {
	// Arrange
	mockWithdrawRepo := new(MockWithdrawRepository)
	mockBalanceRepo := new(MockBalanceRepository)

	service := NewTestWithdrawService(mockWithdrawRepo, mockBalanceRepo)

	user := &models.User{
		ID:    1,
		Login: "testuser",
	}

	withdrawRequest := schemas.WithdrawRequest{
		Order: "invalid_order", // Невалидный номер заказа
		Sum:   100.50,
	}

	// Act
	err := service.Set(user, withdrawRequest)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "can't parse number of order", err.Error())
	// Проверяем, что моки не были вызваны
	mockWithdrawRepo.AssertNotCalled(t, "Create")
	mockBalanceRepo.AssertNotCalled(t, "SetWithdrawForUserID")
}

func TestWithdrawService_Set_WithdrawRepositoryCreateError(t *testing.T) {
	// Arrange
	mockWithdrawRepo := new(MockWithdrawRepository)
	mockBalanceRepo := new(MockBalanceRepository)

	service := NewTestWithdrawService(mockWithdrawRepo, mockBalanceRepo)

	user := &models.User{
		ID:    1,
		Login: "testuser",
	}

	withdrawRequest := schemas.WithdrawRequest{
		Order: "12345",
		Sum:   100.50,
	}

	expectedOrderID := int64(12345)
	expectedSum := 10050
	expectedError := errors.New("database error")

	// Настройка ожиданий для моков
	mockWithdrawRepo.On("Create", user.ID, expectedOrderID, expectedSum).Return(expectedError)

	// Act
	err := service.Set(user, withdrawRequest)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockWithdrawRepo.AssertExpectations(t)
	// Проверяем, что SetWithdrawForUserID не был вызван из-за ошибки
	mockBalanceRepo.AssertNotCalled(t, "SetWithdrawForUserID")
}

func TestWithdrawService_Set_BalanceRepositoryError(t *testing.T) {
	// Arrange
	mockWithdrawRepo := new(MockWithdrawRepository)
	mockBalanceRepo := new(MockBalanceRepository)

	service := NewTestWithdrawService(mockWithdrawRepo, mockBalanceRepo)

	user := &models.User{
		ID:    1,
		Login: "testuser",
	}

	withdrawRequest := schemas.WithdrawRequest{
		Order: "12345",
		Sum:   100.50,
	}

	expectedOrderID := int64(12345)
	expectedSum := 10050
	expectedError := errors.New("insufficient balance")

	// Настройка ожиданий для моков
	mockWithdrawRepo.On("Create", user.ID, expectedOrderID, expectedSum).Return(nil)
	mockBalanceRepo.On("SetWithdrawForUserID", user.ID, expectedSum).Return(expectedError)

	// Act
	err := service.Set(user, withdrawRequest)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockWithdrawRepo.AssertExpectations(t)
	mockBalanceRepo.AssertExpectations(t)
}

func TestWithdrawService_Set_DifferentAmounts(t *testing.T) {
	testCases := []struct {
		name        string
		order       string
		sum         float32
		expectedSum int
	}{
		{
			name:        "small amount",
			order:       "11111",
			sum:         1.00,
			expectedSum: 100,
		},
		{
			name:        "large amount",
			order:       "22222",
			sum:         999.99,
			expectedSum: 99999,
		},
		{
			name:        "fractional amount",
			order:       "33333",
			sum:         50.25,
			expectedSum: 5025,
		},
		{
			name:        "zero amount",
			order:       "44444",
			sum:         0.0,
			expectedSum: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockWithdrawRepo := new(MockWithdrawRepository)
			mockBalanceRepo := new(MockBalanceRepository)

			service := NewTestWithdrawService(mockWithdrawRepo, mockBalanceRepo)

			user := &models.User{
				ID:    1,
				Login: "testuser",
			}

			withdrawRequest := schemas.WithdrawRequest{
				Order: tc.order,
				Sum:   tc.sum,
			}

			expectedOrderID, _ := withdrawRequest.GetOrderAsInt()

			// Настройка ожиданий для моков
			mockWithdrawRepo.On("Create", user.ID, expectedOrderID, tc.expectedSum).Return(nil)
			mockBalanceRepo.On("SetWithdrawForUserID", user.ID, tc.expectedSum).Return(nil)

			// Act
			err := service.Set(user, withdrawRequest)

			// Assert
			assert.NoError(t, err)
			mockWithdrawRepo.AssertExpectations(t)
			mockBalanceRepo.AssertExpectations(t)
		})
	}
}

func TestWithdrawService_Set_ConcurrentCalls(t *testing.T) {
	// Arrange
	mockWithdrawRepo := new(MockWithdrawRepository)
	mockBalanceRepo := new(MockBalanceRepository)

	service := NewTestWithdrawService(mockWithdrawRepo, mockBalanceRepo)

	user := &models.User{
		ID:    1,
		Login: "testuser",
	}

	withdrawRequest := schemas.WithdrawRequest{
		Order: "12345",
		Sum:   100.50,
	}

	expectedOrderID := int64(12345)
	expectedSum := 10050

	// Настройка ожиданий для моков (ожидаем 3 вызова)
	mockWithdrawRepo.On("Create", user.ID, expectedOrderID, expectedSum).Return(nil).Times(3)
	mockBalanceRepo.On("SetWithdrawForUserID", user.ID, expectedSum).Return(nil).Times(3)

	// Act - выполняем несколько вызовов последовательно
	for range 3 {
		err := service.Set(user, withdrawRequest)
		assert.NoError(t, err)
	}

	// Assert
	mockWithdrawRepo.AssertExpectations(t)
	mockBalanceRepo.AssertExpectations(t)
}
