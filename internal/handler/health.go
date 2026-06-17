package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kidkender/prohealth-api/internal/response"
)

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, response.Done())
}
