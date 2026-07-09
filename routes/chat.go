package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kidkender/prohealth-api/internal/handler"
	"github.com/kidkender/prohealth-api/internal/middleware"
)

func RegisterChatRoutes(router gin.IRouter, chatHandler *handler.ChatHandler) {
	router.POST("/chat", middleware.TimeoutMiddleware(30*time.Second), chatHandler.Chat)
}
