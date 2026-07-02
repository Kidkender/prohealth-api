package notification

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/kidkender/prohealth-api/internal/cache"
	"github.com/kidkender/prohealth-api/internal/model"
	"github.com/kidkender/prohealth-api/internal/service"
	"github.com/redis/go-redis/v9"
)

const (
	consumerName  = "consumer-1"
	minIdleTime   = 5 * time.Minute
	reclaimBatch  = 10
	reclaimTicker = 30 * time.Second
)

func RunNotificationConsumer(ctx context.Context, redisClient *redis.Client, mailService *service.MailService) {
	go recoverPendingMessages(ctx, redisClient, mailService)
	consumeNewMessages(ctx, redisClient, mailService)
}

func consumeNewMessages(ctx context.Context, redisClient *redis.Client, mailService *service.MailService) {

	for {
		select {
		case <-ctx.Done():
			slog.Info("notification consumer shutting down")
			return
		default:
		}

		entries, err := redisClient.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    string(cache.NotificationGroup),
			Consumer: consumerName,
			Streams:  []string{string(cache.StreamNotifications), ">"},
			Block:    5 * time.Second,
			Count:    10,
		}).Result()

		if err != nil {
			if err == redis.Nil {
				continue
			}
			slog.Info("xreadgroup error: ", "err", err)

		}
		if len(entries) == 0 {
			continue
		}
		for _, entry := range entries {
			for _, message := range entry.Messages {
				processMessage(ctx, redisClient, mailService, message)
			}
		}
	}
}

func recoverPendingMessages(
	ctx context.Context,
	redisClient *redis.Client,
	mailService *service.MailService,
) {
	ticker := time.NewTicker(reclaimTicker)
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
			entries, nextStart, err := redisClient.XAutoClaim(ctx, &redis.XAutoClaimArgs{
				Stream:   string(cache.StreamNotifications),
				Group:    string(cache.NotificationGroup),
				Consumer: consumerName,
				MinIdle:  minIdleTime,
				Start:    start,
				Count:    reclaimBatch,
			}).Result()
			if err != nil {
				slog.Info("xautoclaim error", "err", err)
				break
			}

			for _, entry := range entries {
				slog.Info("reclaiming abandoned message", "id", entry.ID)
				processMessage(ctx, redisClient, mailService, entry)
			}

			if nextStart == "0" || len(entries) == 0 {
				break
			}
			start = nextStart
		}
	}
}

func processMessage(
	ctx context.Context,
	redisClient *redis.Client,
	mailService *service.MailService,
	entry redis.XMessage,
) {
	payloadStr, ok := entry.Values["payload"].(string)
	if !ok {
		slog.Error("invalid payload, acking to drop poison message", "id", entry.ID)
		redisClient.XAck(ctx, string(cache.StreamNotifications), string(cache.NotificationGroup), entry.ID)
		return
	}

	jobType, _ := entry.Values["type"].(string)
	payload := []byte(payloadStr)

	if err := handleNotification(model.OutboxJobType(jobType), payload, mailService); err != nil {
		slog.Error("handle notification error", "err", err, "id", entry.ID)
		return
	}

	redisClient.XAck(ctx, string(cache.StreamNotifications), string(cache.NotificationGroup), entry.ID)
}

func handleNotification(
	jobType model.OutboxJobType,
	payload []byte,
	mailService *service.MailService,
) error {
	switch jobType {
	case model.OutboxJobTypeAppointmentCreated:
		return handleAppointmentCreated(mailService, payload)
	default:
		slog.Info("Unknow job type")
		return fmt.Errorf("unknown job type: %s", jobType)
	}
}
