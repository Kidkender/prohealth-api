package apperror

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AppError struct {
	Status int
	Code   string
}

func (e AppError) Error() string {
	return e.Code
}

func HandleError(c *gin.Context, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.Status, gin.H{
			"error": appErr.Code,
		})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": "error.internal_server",
	})
}

func BadRequest(code string) *AppError {
	return &AppError{Status: http.StatusBadRequest, Code: code}
}

func InternalServerError(code string) *AppError {
	return &AppError{Status: http.StatusInternalServerError, Code: code}
}

func NotFound(code string) *AppError {
	return &AppError{Status: http.StatusNotFound, Code: code}
}

func Unauthorized(code string) *AppError {
	return &AppError{Status: http.StatusUnauthorized, Code: code}
}

func Forbidden(code string) *AppError {
	return &AppError{Status: http.StatusForbidden, Code: code}
}

func Conflict(code string) *AppError {
	return &AppError{Status: http.StatusConflict, Code: code}
}

func RateLimitExceeded(code string) *AppError {
	return &AppError{Status: http.StatusTooManyRequests, Code: code}
}
