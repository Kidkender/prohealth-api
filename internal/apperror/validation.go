package apperror

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func RegisterValidatorTagName() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" || name == "" {
				return fld.Name
			}
			return name
		})
	}
}

func ValidationError(c *gin.Context, err error) {
	var jsonErr *json.UnmarshalTypeError
	if errors.As(err, &jsonErr) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "error.validate.failed",
			"fields": map[string]string{jsonErr.Field: "invalid_type"},
		})
		return
	}

	var timeErr *time.ParseError
	if errors.As(err, &timeErr) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "error.validate.failed",
			"fields":  map[string]string{"date_field": "invalid_format"},
			"message": "invalid date format, expected: YYYY-MM-DDTHH:MM:SSZ",
		})
		return
	}

	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "error.validate.invalid"})
		return
	}

	fields := make(map[string]string)
	for _, e := range ve {
		fields[e.Field()] = fieldErrorMessage(e)
	}
	c.JSON(http.StatusBadRequest, gin.H{
		"error":  "error.validate.failed",
		"fields": fields,
	})
}

func fieldErrorMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "datetime":
		return fmt.Sprintf("Invalid format, expected: %s", e.Param())
	case "email":
		return "Invalid email format"
	case "oneof":
		return fmt.Sprintf("Must be one of: %s", e.Param())
	case "min":
		return fmt.Sprintf("Must be at least %s", e.Param())
	case "max":
		return fmt.Sprintf("Must be at most %s", e.Param())
	default:
		return fmt.Sprintf("Failed on '%s' validation", e.Tag())
	}
}
