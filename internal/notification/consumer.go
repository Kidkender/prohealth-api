package notification

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/kidkender/prohealth-api/internal/cache"
	"github.com/kidkender/prohealth-api/internal/model"
	"github.com/kidkender/prohealth-api/internal/service"
	"github.com/redis/go-redis/v9"
)

type NotificationConsumer struct {
	redisClient   *redis.Client
	mailService   *service.MailService
	consumerName  string
	minIdleTime   time.Duration
	reclaimBatch  int
	reclaimTicker time.Duration
}

func NewConsumer(
	redisClient *redis.Client,
	mailService *service.MailService,
	consumerName string,
	minIdleTime time.Duration,
	reclaimBatch int,
	reclaimTicker time.Duration,
) *NotificationConsumer {
	return &NotificationConsumer{
		redisClient:   redisClient,
		mailService:   mailService,
		consumerName:  consumerName,
		minIdleTime:   minIdleTime,
		reclaimBatch:  reclaimBatch,
		reclaimTicker: reclaimTicker,
	}
}

// Run blocks until ctx is cancelled and both loops have returned, so no message
// is cut off mid-processing at shutdown.
func (c *NotificationConsumer) Run(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.recoverPendingMessages(ctx)
	}()
	c.consumeNewMessages(ctx)
	wg.Wait()
}

const (
	minReadBackoff = 500 * time.Millisecond
	maxReadBackoff = 30 * time.Second
)

func (c *NotificationConsumer) consumeNewMessages(ctx context.Context) {
	backoff := minReadBackoff

	for {
		select {
		case <-ctx.Done():
			slog.Info("notification consumer shutting down")
			return
		default:
		}

		entries, err := c.redisClient.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    string(cache.NotificationGroup),
			Consumer: c.consumerName,
			Streams:  []string{string(cache.StreamNotifications), ">"},
			Block:    5 * time.Second,
			Count:    10,
		}).Result()

		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			}
			if ctx.Err() != nil {
				continue // shutting down: the select above returns
			}
			// Redis is unreachable or the group is missing: wait before retrying
			// instead of spinning and flooding the log.
			slog.Error("xreadgroup error", "err", err, "retry_in", backoff)
			select {
			case <-ctx.Done():
			case <-time.After(backoff):
			}
			backoff = min(backoff*2, maxReadBackoff)
			continue
		}
		backoff = minReadBackoff
		if len(entries) == 0 {
			continue
		}
		for _, entry := range entries {
			for _, message := range entry.Messages {
				c.processMessage(ctx, message)
			}
		}
	}
}

func (c *NotificationConsumer) recoverPendingMessages(
	ctx context.Context,
) {
	ticker := time.NewTicker(c.reclaimTicker)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("recoverPendingMessages stopped")
			return
		case <-ticker.C:
		}

		start := "0"
		for {
			entries, nextStart, err := c.redisClient.XAutoClaim(ctx, &redis.XAutoClaimArgs{
				Stream:   string(cache.StreamNotifications),
				Group:    string(cache.NotificationGroup),
				Consumer: c.consumerName,
				MinIdle:  c.minIdleTime,
				Start:    start,
				Count:    int64(c.reclaimBatch),
			}).Result()
			if err != nil {
				slog.Info("xautoclaim error", "err", err)
				break
			}

			for _, entry := range entries {
				slog.Info("reclaiming abandoned message", "id", entry.ID)
				c.processMessage(ctx, entry)
			}

			if nextStart == "0" || len(entries) == 0 {
				break
			}
			start = nextStart
		}
	}
}

func (c *NotificationConsumer) processMessage(
	ctx context.Context,
	entry redis.XMessage,
) {
	payloadStr, ok := entry.Values["payload"].(string)
	if !ok {
		slog.Error("invalid payload, acking to drop poison message", "id", entry.ID)
		c.redisClient.XAck(ctx, string(cache.StreamNotifications), string(cache.NotificationGroup), entry.ID)
		return
	}

	jobType, _ := entry.Values["type"].(string)
	payload := []byte(payloadStr)

	if err := handleNotification(ctx, model.OutboxJobType(jobType), payload, c.mailService); err != nil {
		slog.Error("handle notification error", "err", err, "id", entry.ID)
		return
	}

	c.redisClient.XAck(ctx, string(cache.StreamNotifications), string(cache.NotificationGroup), entry.ID)
}

func handleNotification(
	ctx context.Context,
	jobType model.OutboxJobType,
	payload []byte,
	mailService *service.MailService,
) error {
	switch jobType {
	case model.OutboxJobTypeAppointmentCreated:
		return handleAppointmentCreated(ctx, mailService, payload)
	default:
		slog.Info("Unknow job type")
		return fmt.Errorf("unknown job type: %s", jobType)
	}
}
