package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"order-system/internal/service"
)

type HealthController struct {
	healthService service.HealthService
}

func NewHealthController(healthService service.HealthService) *HealthController {
	return &HealthController{healthService: healthService}
}

func (h *HealthController) RegisterRoutes(router gin.IRouter) {
	router.GET("/health", h.GetHealth)
}

func (h *HealthController) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": h.healthService.GetStatus()})
}
