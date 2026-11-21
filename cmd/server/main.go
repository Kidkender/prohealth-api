package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
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

	var workers sync.WaitGroup
	workers.Add(1)
	go func() {
		defer workers.Done()
		c.OutboxWorker.Run(ctx)
	}()
	log.Println("outbox worker started")

	slog.SetDefault(c.Logger)

	r := gin.New()
	r.Use(gin.Recovery())
	r.RedirectTrailingSlash = false
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.AccessLogMiddleware(c.Logger))
	r.Use(middleware.CorsMiddleware(cfg.CORSOrigin))
	r.Use(middleware.AuthMiddleware())
	r.Use(middleware.TenantMiddleware(c.HospitalService))

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

	routes.RegisterRoutes(r, c, authLimiter, apiLimiter, c.RedisClient)
	port := cfg.Port
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{Addr: ":" + port, Handler: r}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()
	log.Println("server listening on :" + port)

	<-ctx.Done()
	log.Println("shutting down: finishing in-flight requests")
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelShutdown()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown: %v", err)
	}
	workers.Wait()
	log.Println("stopped")
}
