package erors

import (
	"errors"
	"fmt"
)

var (
	ErrTrackCodeNotFound  = errors.New("tracking code not found")
	ErrInvalidTrackCode   = errors.New("invalid tracking code format")
	ErrServiceUnavailable = errors.New("tracking service temporarily unavailable")
	ErrRequestTimeout     = errors.New("request timeout")
	ErrTooManyRequests    = errors.New("too many requests, please try again later")
)

var (
	ErrInternalScraping = errors.New("internal scraping error")
	ErrInternalParsing  = errors.New("internal parsing error")
	ErrInternalCache    = errors.New("internal cache error")
	ErrInternalNetwork  = errors.New("internal network error")
	ErrInternalDatabase = errors.New("internal database error")
)

type ErrorType int

const (
	ClientError ErrorType = iota
	InternalError
)

type AppError struct {
	Type    ErrorType
	Message string
	Err     error
	Code    string
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewClientError(message string, err error) *AppError {
	return &AppError{
		Type:    ClientError,
		Message: message,
		Err:     err,
	}
}

func NewInternalError(code, message string, err error) *AppError {
	return &AppError{
		Type:    InternalError,
		Message: message,
		Err:     err,
		Code:    code,
	}
}

func IsClientError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == ClientError
	}
	return false
}

func IsInternalError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == InternalError
	}
	return false
}

func GetErrorCode(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return "UNKNOWN"
}
