package apperrors_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"oengo-api/internal/apperrors"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRequestIDAndStructuredErrorResponse(t *testing.T) {
	r := gin.New()
	r.Use(apperrors.RequestIDMiddleware(), apperrors.ErrorHandlingMiddleware())

	r.GET("/test-not-found", func(c *gin.Context) {
		apperrors.RespondWithError(c, apperrors.NewNotFound("restaurant", "rest_999"))
	})

	r.GET("/test-kyc", func(c *gin.Context) {
		apperrors.RespondWithError(c, apperrors.NewKYCRequired(1, "EU AMLD5 limit exceeded"))
	})

	r.GET("/test-panic", func(c *gin.Context) {
		panic("simulated catastrophic nil pointer dereference")
	})

	t.Run("404 Not Found with custom X-Request-ID propagation", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test-not-found", nil)
		req.Header.Set("X-Request-ID", "req_custom_trace_001")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, "req_custom_trace_001", w.Header().Get("X-Request-ID"))

		var body map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Equal(t, false, body["success"])
		assert.Equal(t, string(apperrors.ErrCodeNotFound), body["code"])
		assert.Equal(t, "restaurant 'rest_999' not found", body["error"])
		assert.Equal(t, "req_custom_trace_001", body["requestId"])
		assert.NotEmpty(t, body["timestamp"])
	})

	t.Run("403 KYC Required with auto-generated Request ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test-kyc", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		reqID := w.Header().Get("X-Request-ID")
		assert.Contains(t, reqID, "req_")

		var body map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Equal(t, string(apperrors.ErrCodeKYCRequired), body["code"])
		details, ok := body["details"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, float64(1), details["requiredKycTier"])
	})

	t.Run("500 Panic Recovery Middleware", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test-panic", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var body map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Equal(t, false, body["success"])
		assert.Equal(t, string(apperrors.ErrCodeInternal), body["code"])
		assert.NotEmpty(t, body["requestId"])
	})
}

func TestWrapDBError(t *testing.T) {
	assert.Nil(t, apperrors.WrapDBError("noop", nil))

	notFoundErr := apperrors.WrapDBError("Order", gorm.ErrRecordNotFound)
	var appErr *apperrors.AppError
	require.ErrorAs(t, notFoundErr, &appErr)
	assert.Equal(t, http.StatusNotFound, appErr.HTTPStatus)
	assert.Equal(t, apperrors.ErrCodeNotFound, appErr.Code)

	dupErr := apperrors.WrapDBError("CreateUser", errors.New("ERROR: duplicate key value violates unique constraint (SQLSTATE 23505)"))
	require.ErrorAs(t, dupErr, &appErr)
	assert.Equal(t, http.StatusConflict, appErr.HTTPStatus)
	assert.Equal(t, apperrors.ErrCodeConflict, appErr.Code)

	checkErr := apperrors.WrapDBError("UpdateWallet", errors.New("ERROR: new row violates check constraint (SQLSTATE 23514)"))
	require.ErrorAs(t, checkErr, &appErr)
	assert.Equal(t, http.StatusBadRequest, appErr.HTTPStatus)
	assert.Equal(t, apperrors.ErrCodeValidation, appErr.Code)
}
