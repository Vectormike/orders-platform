package router

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"order-system/internal/controller"
	"order-system/internal/repository"
	"order-system/internal/service"
)

func New(pool *pgxpool.Pool) (*gin.Engine, error) {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	if err := r.SetTrustedProxies(nil); err != nil {
		return nil, fmt.Errorf("set trusted proxies: %w", err)
	}

	healthRepository := repository.NewHealthRepository()
	healthService := service.NewHealthService(healthRepository)
	healthController := controller.NewHealthController(healthService)
	healthController.RegisterRoutes(r)

	if pool != nil {
		orderRepository := repository.NewOrderRepository(pool)
		orderService := service.NewOrderService(orderRepository)
		orderController := controller.NewOrderController(orderService)
		orderController.RegisterRoutes(r)
	}

	return r, nil
}
