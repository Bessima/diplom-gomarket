package customerror

import (
	"fmt"
	"net/http"
)

type CustomError interface {
	Error() string
	GetHTTPCode() int
}

type UniqueViolationError struct {
	httpCode int
	message  string
}

func NewUniqueViolationError(msg string) *UniqueViolationError {
	return &UniqueViolationError{httpCode: http.StatusUnprocessableEntity, message: msg}
}

func (e *UniqueViolationError) Error() string {
	return fmt.Sprintf("unique violation: %s", e.message)
}

func (e *UniqueViolationError) GetHTTPCode() int {
	return e.httpCode
}

type CommonPGError struct {
	httpCode int
	message  string
}

func NewCommonPGError(msg string) *CommonPGError {
	return &CommonPGError{httpCode: http.StatusInternalServerError, message: msg}
}

func (e *CommonPGError) Error() string {
	return e.message
}

func (e *CommonPGError) GetHTTPCode() int {
	return e.httpCode
}
