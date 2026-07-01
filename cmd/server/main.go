package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kidkender/prohealth-api/internal/apperror"
	"github.com/kidkender/prohealth-api/internal/config"
	"github.com/kidkender/prohealth-api/internal/container"
	"github.com/kidkender/prohealth-api/internal/middleware"
	"github.com/kidkender/prohealth-api/internal/ratelimit"
	"github.com/kidkender/prohealth-api/pkg/database"
	"github.com/kidkender/prohealth-api/routes"
)

func main() {
	cfg := config.LoadEnv()
	apperror.RegisterValidatorTagName()

	db, err := database.NewPostgresDB()
	if err != nil {
		panic(err)
	}

	c, err := container.New(db, cfg)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.Println("outbox worker started")
	go c.OutboxWorker.Run(ctx)

	r := gin.Default()
	r.RedirectTrailingSlash = false
	r.Use(middleware.CorsMiddleware(cfg.CORSOrigin))
	r.Use(middleware.AuthMiddleware())
	r.Use(middleware.TenantMiddleware(db))

	authLimiter, err := ratelimit.NewTokenBucket(c.RedisClient, ratelimit.Config{
		Capacity:       10,
		RefillRate:     10,
		RefillInterval: time.Minute,
	})
	if err != nil {
		log.Fatal(err)
	}
	apiLimiter, err := ratelimit.NewTokenBucket(c.RedisClient, ratelimit.Config{
		Capacity:       200,
		RefillRate:     200,
		RefillInterval: time.Minute,
	})
	if err != nil {
		log.Fatal(err)
	}

	routes.RegisterRoutes(r, c, authLimiter, apiLimiter)
	port := cfg.Port
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
