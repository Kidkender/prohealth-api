package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDKey = "request_id"
const RequestIDHeader = "X-Request-ID"

type requestIDCtxKey struct{}

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader(RequestIDHeader)
		if reqID == "" {
			reqID = uuid.NewString()
		}
		c.Set(RequestIDKey, reqID)
		ctx := context.WithValue(c.Request.Context(), requestIDCtxKey{}, reqID)
		c.Request = c.Request.WithContext(ctx)
		c.Writer.Header().Set(RequestIDHeader, reqID)
		c.Next()
	}
}

func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDCtxKey{}).(string); ok {
		return v
	}
	return ""
}
