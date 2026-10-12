package apperrors

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"
)

// ErrorCode represents standardized machine-readable FinTech & API error codes.
type ErrorCode string

const (
	ErrCodeBadRequest          ErrorCode = "ERR_BAD_REQUEST"
	ErrCodeValidation          ErrorCode = "ERR_VALIDATION_FAILED"
	ErrCodeUnauthorized        ErrorCode = "ERR_UNAUTHORIZED"
	ErrCodeForbidden           ErrorCode = "ERR_FORBIDDEN"
	ErrCodeKYCRequired         ErrorCode = "ERR_KYC_VERIFICATION_REQUIRED"
	ErrCodeNotFound            ErrorCode = "ERR_RESOURCE_NOT_FOUND"
	ErrCodeConflict            ErrorCode = "ERR_CONFLICT_DUPLICATE"
	ErrCodeInsufficientFunds   ErrorCode = "ERR_INSUFFICIENT_FUNDS"
	ErrCodeLedgerImbalance     ErrorCode = "ERR_LEDGER_IMBALANCE"
	ErrCodeEscrowLocked        ErrorCode = "ERR_ESCROW_LOCKED"
	ErrCodeInvalidSecurityCode ErrorCode = "ERR_INVALID_SECURITY_CODE"
	ErrCodeDatabase            ErrorCode = "ERR_DATABASE_OPERATION"
	ErrCodeInternal            ErrorCode = "ERR_INTERNAL_SERVER"
)

// AppError is an RFC-7807 inspired structured error for enterprise banking & commerce APIs.
type AppError struct {
	HTTPStatus int                    `json:"status"`
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	RequestID  string                 `json:"requestId,omitempty"`
	Timestamp  string                 `json:"timestamp"`
	Cause      error                  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

func nowUTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func NewBadRequest(message string, details ...map[string]interface{}) *AppError {
	var det map[string]interface{}
	if len(details) > 0 {
		det = details[0]
	}
	return &AppError{
		HTTPStatus: http.StatusBadRequest,
		Code:       ErrCodeBadRequest,
		Message:    message,
		Details:    det,
		Timestamp:  nowUTC(),
	}
}

func NewValidation(field, reason string) *AppError {
	return &AppError{
		HTTPStatus: http.StatusUnprocessableEntity,
		Code:       ErrCodeValidation,
		Message:    fmt.Sprintf("Validation failed on field '%s': %s", field, reason),
		Details: map[string]interface{}{
			"field":  field,
			"reason": reason,
		},
		Timestamp: nowUTC(),
	}
}

func NewUnauthorized(message string) *AppError {
	if message == "" {
		message = "Authentication required or invalid credentials"
	}
	return &AppError{
		HTTPStatus: http.StatusUnauthorized,
		Code:       ErrCodeUnauthorized,
		Message:    message,
		Timestamp:  nowUTC(),
	}
}

func NewForbidden(message string) *AppError {
	if message == "" {
		message = "You do not have permission to perform this operation"
	}
	return &AppError{
		HTTPStatus: http.StatusForbidden,
		Code:       ErrCodeForbidden,
		Message:    message,
		Timestamp:  nowUTC(),
	}
}

func NewKYCRequired(requiredTier int, message string) *AppError {
	return &AppError{
		HTTPStatus: http.StatusForbidden,
		Code:       ErrCodeKYCRequired,
		Message:    message,
		Details: map[string]interface{}{
			"requiredKycTier": requiredTier,
			"regulation":      "EU AMLD5 / MiCA Compliance",
		},
		Timestamp: nowUTC(),
	}
}

func NewNotFound(resource, identifier string) *AppError {
	msg := fmt.Sprintf("%s not found", resource)
	if identifier != "" {
		msg = fmt.Sprintf("%s '%s' not found", resource, identifier)
	}
	return &AppError{
		HTTPStatus: http.StatusNotFound,
		Code:       ErrCodeNotFound,
		Message:    msg,
		Details: map[string]interface{}{
			"resource":   resource,
			"identifier": identifier,
		},
		Timestamp: nowUTC(),
	}
}

func NewConflict(message string) *AppError {
	return &AppError{
		HTTPStatus: http.StatusConflict,
		Code:       ErrCodeConflict,
		Message:    message,
		Timestamp:  nowUTC(),
	}
}

func NewInsufficientFunds(available, required float64, currency string) *AppError {
	return &AppError{
		HTTPStatus: http.StatusPaymentRequired,
		Code:       ErrCodeInsufficientFunds,
		Message:    fmt.Sprintf("insufficient %s balance: available %.2f, required %.2f", currency, available, required),
		Details: map[string]interface{}{
			"available": available,
			"required":  required,
			"currency":  currency,
		},
		Timestamp: nowUTC(),
	}
}

func NewInvalidSecurityCode(message string) *AppError {
	return &AppError{
		HTTPStatus: http.StatusBadRequest,
		Code:       ErrCodeInvalidSecurityCode,
		Message:    message,
		Timestamp:  nowUTC(),
	}
}

func NewDatabaseError(operation string, err error) *AppError {
	return &AppError{
		HTTPStatus: http.StatusInternalServerError,
		Code:       ErrCodeDatabase,
		Message:    fmt.Sprintf("Database error during %s", operation),
		Details: map[string]interface{}{
			"operation": operation,
		},
		Timestamp: nowUTC(),
		Cause:     err,
	}
}

func NewInternal(message string, cause error) *AppError {
	if message == "" {
		message = "An unexpected internal server error occurred"
	}
	return &AppError{
		HTTPStatus: http.StatusInternalServerError,
		Code:       ErrCodeInternal,
		Message:    message,
		Timestamp:  nowUTC(),
		Cause:     cause,
	}
}

// WrapDBError translates GORM and PostgreSQL driver errors into domain AppErrors.
func WrapDBError(operation string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return NewNotFound(operation, "")
	}
	errMsg := strings.ToLower(err.Error())
	if strings.Contains(errMsg, "duplicate key") || strings.Contains(errMsg, "unique constraint") || strings.Contains(errMsg, "23505") {
		return &AppError{
			HTTPStatus: http.StatusConflict,
			Code:       ErrCodeConflict,
			Message:    fmt.Sprintf("Duplicate record conflict during %s", operation),
			Timestamp:  nowUTC(),
			Cause:     err,
		}
	}
	if strings.Contains(errMsg, "check constraint") || strings.Contains(errMsg, "23514") {
		return &AppError{
			HTTPStatus: http.StatusBadRequest,
			Code:       ErrCodeValidation,
			Message:    fmt.Sprintf("Database constraint violation during %s", operation),
			Timestamp:  nowUTC(),
			Cause:     err,
		}
	}
	return NewDatabaseError(operation, err)
}
