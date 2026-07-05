package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kidkender/prohealth-api/internal/cache"
	"github.com/redis/go-redis/v9"
)

type IdempotencyStatus string

const (
	statusProcessing IdempotencyStatus = "processing"
	statusCompleted  IdempotencyStatus = "completed"
)

type idempotencyRecord struct {
	Status      IdempotencyStatus
	RequestHash string
	StatusCode  int
	Headers     map[string][]string
	Body        json.RawMessage
}

type bodyCapture struct {
	gin.ResponseWriter
	buf *bytes.Buffer
}

const maxIdempotencyBodySize = 1 << 20
const processingTTL = 15 * time.Second
const completeTTL = 30 * time.Minute

func (b *bodyCapture) Write(p []byte) (int, error) {
	b.buf.Write(p)
	return b.ResponseWriter.Write(p)
}

func requestScope(c *gin.Context) string {
	if v, ok := c.Get("user_id"); ok {
		if userID, ok := v.(uint); ok {
			return fmt.Sprintf("user:%d", userID)
		}
	}
	return "ip:" + c.ClientIP()
}

func tryRelay(c *gin.Context, redisClient *redis.Client, redisKey, requestHash string) bool {
	raw, err := redisClient.Get(c.Request.Context(), redisKey).Result()
	if err == redis.Nil {
		return false
	}

	if err != nil {
		slog.Error("idempotency redis error: ", "error", err)
		return false
	}

	var record idempotencyRecord

	if json.Unmarshal([]byte(raw), &record) != nil {
		return false
	}

	if record.RequestHash != requestHash {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{
			"error": "error.idempotency.key_reused_differenet",
		})
		return true
	}

	if record.Status == statusProcessing {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{
			"error": "error.idempotency.processing",
		})
		return true
	}

	for k, values := range record.Headers {
		for _, v := range values {
			c.Header(k, v)
		}
	}
	c.Data(record.StatusCode, "application/json", record.Body)
	c.Abort()
	return true
}

func IdempotencyMiddleware(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch:

		default:
			c.Next()
			return
		}
		key := c.GetHeader("Idempotency-Key")
		if key == "" {
			c.Next()
			return
		}

		rawBody, err := io.ReadAll(io.LimitReader(c.Request.Body, maxIdempotencyBodySize))
		if err != nil {
			c.Next()
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewReader(rawBody))
		hash := sha256.Sum256(rawBody)
		requestHash := hex.EncodeToString(hash[:])

		scope := requestScope(c)
		redisKey := cache.IdempotencyKey(scope, c.FullPath(), key)
		ctx := c.Request.Context()

		if handled := tryRelay(c, redisClient, redisKey, requestHash); handled {
			return
		}

		placeholder, _ := json.Marshal(idempotencyRecord{
			Status:      statusProcessing,
			RequestHash: requestHash,
		})
		ok, err := redisClient.SetNX(ctx, redisKey, placeholder, processingTTL).Result()
		if err != nil {
			slog.Error("idempotency setnx error: ", "err", err)
			c.Next()
			return
		}

		if !ok {
			if handled := tryRelay(c, redisClient, redisKey, requestHash); handled {
				return
			}
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{
				"error": "error.idempotency.processing",
			})
			return
		}

		writer := &bodyCapture{ResponseWriter: c.Writer, buf: &bytes.Buffer{}}
		c.Writer = writer
		c.Next()

		record := idempotencyRecord{
			Status:      statusCompleted,
			RequestHash: requestHash,
			StatusCode:  writer.Status(),
			Headers:     writer.Header(),
			Body:        writer.buf.Bytes(),
		}

		if writer.Status() >= http.StatusInternalServerError {
			redisClient.Del(ctx, redisKey)
			return
		}

		data, _ := json.Marshal(record)
		if err := redisClient.Set(ctx, redisKey, data, completeTTL).Err(); err != nil {
			slog.Error("idempotency set error: ", "err", err)
			redisClient.Del(ctx, redisKey)
		}

	}
}
