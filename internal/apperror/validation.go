package apperror

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func ValidationError(c *gin.Context, err error) {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "error.request.invalid",
		})
		return
	}

	fields := make(map[string]string)
	for _, e := range ve {
		fields[strings.ToLower(e.Field())] = e.Tag()
	}
	c.JSON(http.StatusBadRequest, gin.H{
		"error":  "error.validation.failed",
		"fields": fields,
	})
}
