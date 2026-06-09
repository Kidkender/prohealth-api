package main

import (
	"github.com/gin-gonic/gin"
	"github.com/kidkender/prohealth-api/internal/config"
	"github.com/kidkender/prohealth-api/internal/middleware"
	"github.com/kidkender/prohealth-api/pkg/database"
	"github.com/kidkender/prohealth-api/routes"
)

func main() {
	config.LoadEnv()
	db, err := database.NewPostgresDB()
	if err != nil {
		panic(err)
	}

	r := gin.Default()

	routes.RegisterRoutes(r, db)
	r.Use(middleware.AuthMiddleware())
	r.Run(":8080")
}
