package utils

import "net/http"

type AppError struct {
	Code    int
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

func NewClientError(msg string) error {
	return &AppError{Code: http.StatusBadRequest, Message: msg}
}

func NewNotFoundError(msg string) error {
	return &AppError{Code: http.StatusNotFound, Message: msg}
}
