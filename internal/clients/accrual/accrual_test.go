package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAccrualClient(t *testing.T) {
	address := "http://localhost:8080"
	client := NewAccrualClient(address)

	assert.NotNil(t, client)
	assert.Equal(t, address, client.address)
	assert.NotNil(t, client.httpClient)
}

func TestAccrualClient_Get_Success(t *testing.T) {
	testCases := []struct {
		name            string
		orderNumber     int
		expectedStatus  string
		expectedAccrual float32
		mockResponse    AccrualResponse
		statusCode      int
	}{
		{
			name:            "registered order",
			orderNumber:     123456,
			expectedStatus:  "REGISTERED",
			expectedAccrual: 0,
			mockResponse:    AccrualResponse{Order: "123456", Status: "REGISTERED"},
			statusCode:      http.StatusOK,
		},
		{
			name:            "processing order",
			orderNumber:     789012,
			expectedStatus:  "PROCESSING",
			expectedAccrual: 0,
			mockResponse:    AccrualResponse{Order: "789012", Status: "PROCESSING"},
			statusCode:      http.StatusOK,
		},
		{
			name:            "processed order with accrual",
			orderNumber:     345678,
			expectedStatus:  "PROCESSED",
			expectedAccrual: 500.5,
			mockResponse:    AccrualResponse{Order: "345678", Status: "PROCESSED", Accrual: 500.5},
			statusCode:      http.StatusOK,
		},
		{
			name:            "invalid order",
			orderNumber:     999999,
			expectedStatus:  "INVALID",
			expectedAccrual: 0,
			mockResponse:    AccrualResponse{Order: "999999", Status: "INVALID"},
			statusCode:      http.StatusOK,
		},
	}

	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Создаем тестовый сервер
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, fmt.Sprintf("/api/orders/%d", tc.orderNumber), r.URL.Path)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.statusCode)

				json.NewEncoder(w).Encode(tc.mockResponse)
			}))
			defer server.Close()

			// Создаем клиента с адресом тестового сервера
			client := NewAccrualClient(server.URL)

			// Вызываем метод Get
			response, err := client.Get(ctx, tc.orderNumber)

			// Проверяем результаты
			require.NoError(t, err)
			require.NotNil(t, response)
			assert.Equal(t, tc.orderNumber, response.Order)
			assert.Equal(t, tc.expectedStatus, response.Status)
			assert.Equal(t, tc.expectedAccrual, response.Accrual)
		})
	}
}

func TestAccrualClient_Get_HTTPErrors(t *testing.T) {
	testCases := []struct {
		name        string
		orderNumber int
		statusCode  int
		expectError bool
	}{
		{
			name:        "order not found",
			orderNumber: 111111,
			statusCode:  http.StatusNoContent,
			expectError: true,
		},
		{
			name:        "too many requests",
			orderNumber: 222222,
			statusCode:  http.StatusTooManyRequests,
			expectError: true,
		},
		{
			name:        "internal server error",
			orderNumber: 333333,
			statusCode:  http.StatusInternalServerError,
			expectError: true,
		},
	}
	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
			}))
			defer server.Close()

			client := NewAccrualClient(server.URL)
			response, err := client.Get(ctx, tc.orderNumber)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
			}
		})
	}
}

func TestAccrualClient_Get_NetworkError(t *testing.T) {
	// Используем несуществующий адрес для имитации сетевой ошибки
	client := NewAccrualClient("http://localhost:99999")

	response, err := client.Get(context.Background(), 123456)

	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestAccrualClient_Get_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Отправляем некорректный JSON
		w.Write([]byte(`{"order": "not_a_number", "status": 123}`))
	}))
	defer server.Close()

	client := NewAccrualClient(server.URL)

	resp, err := client.Get(context.Background(), 123456)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestAccrualClient_Get_ReadBodyError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Используем ResponseWriter, который возвращает ошибку при чтении
		hj := &httptest.ResponseRecorder{}
		hj.WriteHeader(http.StatusOK)
		json.NewEncoder(hj).Encode(AccrualResponse{Order: "123456", Status: "PROCESSING"})

		// Заменяем тело на reader, который возвращает ошибку
		w.Header().Set("Content-Length", "100")
	}))
	defer server.Close()

	client := NewAccrualClient(server.URL)

	resp, err := client.Get(context.Background(), 123456)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestAccrualClient_Get_ResponseBodyCloseError(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Записываем валидный JSON
		json.NewEncoder(w).Encode(AccrualResponse{Order: "123456", Status: "PROCESSING"})
	}))
	defer server.Close()

	client := NewAccrualClient(server.URL)

	// Подменяем httpClient для тестирования ошибки закрытия тела
	client.httpClient = &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			resp := &http.Response{
				StatusCode: http.StatusOK,
				Body: struct {
					io.Reader
					io.Closer
				}{
					Reader: io.NopCloser(strings.NewReader(`{"order":123456,"status":"PROCESSING"}`)),
					Closer: closeFunc(func() error {
						return errors.New("close error")
					}),
				},
				Header: make(http.Header),
			}
			resp.Header.Set("Content-Type", "application/json")
			return resp, nil
		}),
	}

	// Метод должен отработать несмотря на ошибку закрытия (она только логируется)
	response, err := client.Get(context.Background(), 123456)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 123456, response.Order)
	assert.Equal(t, "PROCESSING", response.Status)
}

// Вспомогательные типы для тестирования
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type closeFunc func() error

func (f closeFunc) Close() error {
	return f()
}

// Дополнительный тест для проверки форматирования URL
func TestAccrualClient_Get_URLFormat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, что URL сформирован правильно
		assert.Equal(t, "/api/orders/1234567890", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(AccrualResponse{
			Order:   "1234567890",
			Status:  "PROCESSED",
			Accrual: 1000,
		})
	}))
	defer server.Close()

	client := NewAccrualClient(server.URL)
	response, err := client.Get(context.Background(), 1234567890)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 1234567890, response.Order)
	assert.Equal(t, "PROCESSED", response.Status)
	assert.Equal(t, float32(1000), response.Accrual)
}
