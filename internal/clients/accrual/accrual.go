package accrual

import (
	"encoding/json"
	"fmt"
	"github.com/Bessima/diplom-gomarket/internal/middlewares/logger"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
)

type AccrualResponse struct {
	Order   int     `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual,omitempty"`
}

type AccrualClient struct {
	httpClient *http.Client
	address    string
}

func NewAccrualClient(address string) *AccrualClient {
	client := AccrualClient{address: address, httpClient: &http.Client{}}
	return &client
}

func (client AccrualClient) Get(orderNumber int) (*AccrualResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%d", client.address, orderNumber)
	response, err := client.httpClient.Get(url)
	if err != nil {
		log.Printf("Failed to create resource at: %s and the error is: %v\n", url, err)
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		err := fmt.Errorf("failed to create resource at: %s , answer was with status code %d", url, response.StatusCode)
		return nil, err
	}

	defer func() {
		if err := response.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v\n", err)
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
}
