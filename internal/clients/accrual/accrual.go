package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Bessima/diplom-gomarket/internal/middlewares/logger"
	"github.com/Bessima/diplom-gomarket/internal/retry"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
	"time"
)

const timeToSleep = time.Duration(30) * time.Second

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
}

func NewAccrualClient(address string) *AccrualClient {
	client := AccrualClient{address: address, httpClient: &http.Client{}}
	return &client
}

func (client AccrualClient) Get(ctx context.Context, orderNumber string) (*AccrualResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%s", client.address, orderNumber)

	return retry.DoRetryWithResult(ctx, func() (*AccrualResponse, error) {
		response, err := client.httpClient.Get(url)

		if err != nil {
			err = fmt.Errorf("failed to create resource at: %s and the error is: %w", url, err)
			return nil, err
		}

		if response.StatusCode != http.StatusOK {
			if response.StatusCode == http.StatusTooManyRequests {
				err := fmt.Errorf("failed to create resource at: %s , too many requests by accrual system", url)
				time.Sleep(timeToSleep)
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
