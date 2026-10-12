package apperrors

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDKey = "X-Request-ID"

// RequestIDMiddleware attaches a unique correlation ID to every request and response.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader(RequestIDKey)
		if reqID == "" {
			reqID = fmt.Sprintf("req_%s", strings.ReplaceAll(uuid.New().String()[:12], "-", ""))
		}
		c.Set(RequestIDKey, reqID)
		c.Writer.Header().Set(RequestIDKey, reqID)
		c.Next()
	}
}

// GetRequestID retrieves the correlation ID from the Gin context.
func GetRequestID(c *gin.Context) string {
	if val, exists := c.Get(RequestIDKey); exists {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// ErrorHandlingMiddleware provides centralized panic recovery and c.Errors inspection.
func ErrorHandlingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				reqID := GetRequestID(c)
				stack := string(debug.Stack())
				log.Printf("[PANIC RECOVERED] [RequestID=%s] panic=%v\n%s", reqID, rec, stack)

				appErr := NewInternal("An unexpected internal server error occurred", fmt.Errorf("%v", rec))
				appErr.RequestID = reqID

				c.AbortWithStatusJSON(appErr.HTTPStatus, gin.H{
					"success":   false,
					"error":     appErr.Message,
					"code":      appErr.Code,
					"requestId": appErr.RequestID,
					"timestamp": appErr.Timestamp,
				})
			}
		}()

		c.Next()

		// If handler pushed errors to c.Errors and hasn't written a response yet
		if len(c.Errors) > 0 && !c.Writer.Written() {
			RespondWithError(c, c.Errors.Last().Err)
		}
	}
}

// ClassifyError converts any Go error into a structured *AppError with accurate HTTP status and ErrorCode.
func ClassifyError(err error) *AppError {
	if err == nil {
		return nil
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		if appErr.Timestamp == "" {
			appErr.Timestamp = nowUTC()
		}
		if appErr.HTTPStatus == 0 {
			appErr.HTTPStatus = http.StatusBadRequest
		}
		return appErr
	}

	msg := err.Error()
	msgLower := strings.ToLower(msg)

	switch {
	case strings.Contains(msgLower, "not found"):
		return &AppError{
			HTTPStatus: http.StatusNotFound,
			Code:       ErrCodeNotFound,
			Message:    msg,
			Timestamp:  nowUTC(),
			Cause:      err,
		}
	case strings.Contains(msgLower, "invalid credentials") || strings.Contains(msgLower, "unauthorized"):
		return &AppError{
			HTTPStatus: http.StatusUnauthorized,
			Code:       ErrCodeUnauthorized,
			Message:    msg,
			Timestamp:  nowUTC(),
			Cause:      err,
		}
	case strings.Contains(msgLower, "already exists") || strings.Contains(msgLower, "duplicate"):
		return &AppError{
			HTTPStatus: http.StatusConflict,
			Code:       ErrCodeConflict,
			Message:    msg,
			Timestamp:  nowUTC(),
			Cause:      err,
		}
	case strings.Contains(msgLower, "insufficient"):
		return &AppError{
			HTTPStatus: http.StatusBadRequest,
			Code:       ErrCodeInsufficientFunds,
			Message:    msg,
			Timestamp:  nowUTC(),
			Cause:      err,
		}
	case strings.Contains(msgLower, "kyc"):
		return &AppError{
			HTTPStatus: http.StatusForbidden,
			Code:       ErrCodeKYCRequired,
			Message:    msg,
			Timestamp:  nowUTC(),
			Cause:      err,
		}
	case strings.Contains(msgLower, "pin") || strings.Contains(msgLower, "barcode"):
		return &AppError{
			HTTPStatus: http.StatusBadRequest,
			Code:       ErrCodeInvalidSecurityCode,
			Message:    msg,
			Timestamp:  nowUTC(),
			Cause:      err,
		}
	default:
		return &AppError{
			HTTPStatus: http.StatusBadRequest,
			Code:       ErrCodeBadRequest,
			Message:    msg,
			Timestamp:  nowUTC(),
			Cause:      err,
		}
	}
}

// RespondWithError serializes any error into a standardized FinTech JSON response with RequestID and HTTP status.
func RespondWithError(c *gin.Context, err error, overrideStatus ...int) {
	appErr := ClassifyError(err)
	if appErr == nil {
		return
	}

	reqID := GetRequestID(c)
	appErr.RequestID = reqID

	status := appErr.HTTPStatus
	if len(overrideStatus) > 0 && overrideStatus[0] > 0 {
		status = overrideStatus[0]
	}

	if status >= 500 {
		log.Printf("[ERROR 5XX] [RequestID=%s] [Code=%s] %s (cause=%v)", reqID, appErr.Code, appErr.Message, appErr.Cause)
	}

	payload := gin.H{
		"success":   false,
		"error":     appErr.Message,
		"code":      appErr.Code,
		"requestId": appErr.RequestID,
		"timestamp": appErr.Timestamp,
	}
	if len(appErr.Details) > 0 {
		payload["details"] = appErr.Details
	}

	c.JSON(status, payload)
}
