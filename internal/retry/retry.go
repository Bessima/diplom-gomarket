package retry

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"log"
	"time"
)

// RetryConfig конфигурация для повторных попыток
type RetryConfig struct {
	MaxRetries  int
	Delays      []time.Duration
	ShouldRetry func(error) bool
}

var AccrualRetryConfig = RetryConfig{
	MaxRetries: 3,
	Delays:     []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
	ShouldRetry: func(err error) bool {
		return true
	},
}

var PostgresStorageRetryConfig = RetryConfig{
	MaxRetries: 3,
	Delays:     []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
	ShouldRetry: func(err error) bool {
		return IsConnectionExceptionPG(err)
	},
}

func IsConnectionExceptionPG(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		// Класс 08 - Ошибки соединения
		case pgerrcode.ConnectionException,
			pgerrcode.ConnectionDoesNotExist,
			pgerrcode.ConnectionFailure,
			pgerrcode.SQLClientUnableToEstablishSQLConnection,
			pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection,
			pgerrcode.TransactionResolutionUnknown,
			pgerrcode.ProtocolViolation:
			return true
		default:
			return false
		}
	}
	return false
}

// DoRetry выполняет функцию с повторными попытками при ошибках соединения
func DoRetry(ctx context.Context, fn func() error, config ...RetryConfig) error {
	cfg := PostgresStorageRetryConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	var lastErr error

	for attempt := 0; attempt < cfg.MaxRetries; attempt++ {
		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		// Проверяем, нужно ли повторять для этой ошибки
		if !cfg.ShouldRetry(lastErr) {
			return lastErr
		}

		log.Printf("DoRetry attempt %d failed: %v\n", attempt+1, lastErr)

		// Выбираем задержку для текущей попытки
		delay := getDelay(cfg.Delays, attempt)
		if delay > 0 {
			time.Sleep(delay)
		}
	}

	return fmt.Errorf("after %d retries operation failed, last error: %v", cfg.MaxRetries, lastErr)
}

// DoRetryWithResult выполняет функцию с возвращаемым значением и повторными попытками
func DoRetryWithResult[T any](ctx context.Context, fn func() (T, error), config ...RetryConfig) (T, error) {
	var zero T
	cfg := PostgresStorageRetryConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	var lastErr error
	var result T

	for attempt := 0; attempt < cfg.MaxRetries; attempt++ {
		result, lastErr = fn()
		if lastErr == nil {
			return result, nil
		}

		// Проверяем, нужно ли повторять для этой ошибки
		if !cfg.ShouldRetry(lastErr) {
			return zero, lastErr
		}

		log.Printf("DoRetry attempt %d failed: %v\n", attempt+1, lastErr)

		// Выбираем задержку для текущей попытки
		delay := getDelay(cfg.Delays, attempt)
		if delay > 0 {
			time.Sleep(delay)
		}
	}

	return zero, fmt.Errorf("after %d retries operation failed, last error: %v", cfg.MaxRetries, lastErr)
}

// getDelay возвращает задержку для текущей попытки
func getDelay(delays []time.Duration, attempt int) time.Duration {
	if attempt < len(delays) {
		return delays[attempt]
	}
	// Если попыток больше, чем заданных задержек, используем последнюю
	if len(delays) > 0 {
		return delays[len(delays)-1]
	}
	return 0
}
