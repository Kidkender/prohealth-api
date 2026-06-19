package main

import (
	"github.com/gin-gonic/gin"
	"github.com/kidkender/prohealth-api/internal/apperror"
	"github.com/kidkender/prohealth-api/internal/config"
	"github.com/kidkender/prohealth-api/internal/container"
	"github.com/kidkender/prohealth-api/internal/middleware"
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

	c := container.New(db, cfg)

	r := gin.Default()
	r.Use(middleware.CorsMiddleware(cfg.CORSOrigin))
	r.Use(middleware.AuthMiddleware())

	routes.RegisterRoutes(r, c)
	port := cfg.Port
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
