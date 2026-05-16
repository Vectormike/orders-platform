package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"order-system/internal/service"
)

type OrderController struct {
	orderService service.OrderService
}

type createOrderRequest struct {
	CustomerName string `json:"customer_name"`
	AmountCents  int64  `json:"amount_cents"`
}

func NewOrderController(orderService service.OrderService) *OrderController {
	return &OrderController{orderService: orderService}
}

func (o *OrderController) RegisterRoutes(router gin.IRouter) {
	router.POST("/orders", o.CreateOrder)
	router.GET("/orders/:id", o.GetOrder)
}

func (o *OrderController) CreateOrder(c *gin.Context) {
	var request createOrderRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	order, err := o.orderService.CreateOrder(c.Request.Context(), service.CreateOrderInput{
		CustomerName: request.CustomerName,
		AmountCents:  request.AmountCents,
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidOrderInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order"})
		return
	}

	c.JSON(http.StatusCreated, order)
}

func (o *OrderController) GetOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	order, err := o.orderService.GetOrder(c.Request.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidOrderInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrOrderNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch order"})
		}
		return
	}

	c.JSON(http.StatusOK, order)
}
