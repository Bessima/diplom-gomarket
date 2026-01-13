package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/Bessima/diplom-gomarket/internal/middlewares/logger"
	"github.com/Bessima/diplom-gomarket/internal/retry"
	"go.uber.org/zap"
)

const (
	defaultRetryDelay = 30 * time.Second
)

// Атомарная переменная для хранения времени следующего разрешенного запроса
var nextAllowedRequestTime atomic.Int64

type AccrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual,omitempty"`
}

type AccrualClientI interface {
	Get(ctx context.Context, orderID string) (*AccrualResponse, error)
}

type AccrualClient struct {
	httpClient *http.Client
	address    string

	retryDelay time.Duration
}

func NewAccrualClient(address string) *AccrualClient {
	client := AccrualClient{
		address:    address,
		httpClient: &http.Client{},
		retryDelay: defaultRetryDelay,
	}
	return &client
}

func (client AccrualClient) Get(ctx context.Context, orderNumber string) (*AccrualResponse, error) {

	if err := waitIfNeeded(ctx); err != nil {
		return nil, fmt.Errorf("waiting was interrupted: %w", err)
	}

	url := fmt.Sprintf("%s/api/orders/%s", client.address, orderNumber)

	return retry.DoRetryWithResult(ctx, func() (*AccrualResponse, error) {

		if err := waitIfNeeded(ctx); err != nil {
			return nil, fmt.Errorf("waiting was interrupted: %w", err)
		}

		response, err := client.httpClient.Get(url)

		if err != nil {
			err = fmt.Errorf("failed to create resource at: %s and the error is: %w", url, err)
			return nil, err
		}

		if response.StatusCode != http.StatusOK {
			if response.StatusCode == http.StatusTooManyRequests {
				setNextAllowedTime(client.retryDelay)

				err := fmt.Errorf("failed to create resource at: %s , too many requests by accrual system, retry after: %v",
					url, client.retryDelay)
				return nil, err
			}
			err := fmt.Errorf("failed to create resource at: %s , answer was with status code %d", url, response.StatusCode)
			return nil, err
		}

		defer func() {
			if err := response.Body.Close(); err != nil {
				customErr := fmt.Errorf("error closing response body: %v", err)
				logger.Log.Warn(customErr.Error())
			}
		}()

		body, err := io.ReadAll(response.Body)
		if err != nil {
			logger.Log.Error("Error reading response body", zap.Error(err))
			return nil, err
		}

		var answer AccrualResponse
		err = json.Unmarshal(body, &answer)
		if err != nil {
			logger.Log.Error("Error unmarshalling JSON", zap.Error(err))
			return nil, err
		}

		log.Println("Successful getting answer for order: ", orderNumber)

		return &answer, nil
	}, retry.AccrualRetryConfig)
}

func waitIfNeeded(ctx context.Context) error {
	nextTime := nextAllowedRequestTime.Load()
	if nextTime == 0 {
		return nil
	}

	now := time.Now().UnixNano()
	if now < nextTime {
		waitDuration := time.Duration(nextTime - now)
		logger.Log.Debug("Waiting due to rate limiting",
			zap.Duration("wait", waitDuration),
			zap.Time("until", time.Unix(0, nextTime)))

		timer := time.NewTimer(waitDuration)
		defer timer.Stop()

		select {
		case <-timer.C:
			return nil
		case <-ctx.Done():
			logger.Log.Info("Wait cancelled by context", zap.Error(ctx.Err()))
			return ctx.Err()
		}
	}

	return nil
}

func setNextAllowedTime(waitDuration time.Duration) {
	nextTime := time.Now().Add(waitDuration).UnixNano()
	nextAllowedRequestTime.Store(nextTime)

	logger.Log.Info("Rate limit encountered, delaying next request",
		zap.Duration("wait_duration", waitDuration),
		zap.Time("next_allowed", time.Unix(0, nextTime)))
}

func ResetNextAllowedTime() {
	nextAllowedRequestTime.Store(0)
	logger.Log.Debug("Rate limit timer reset")
}
