package service

import (
	"context"
	"errors"
	"strings"

	"order-system/internal/model"
	"order-system/internal/repository"
)

var ErrInvalidOrderInput = errors.New("invalid order input")
var ErrOrderNotFound = errors.New("order not found")

type CreateOrderInput struct {
	CustomerName string
	AmountCents  int64
}

type OrderService interface {
	CreateOrder(ctx context.Context, input CreateOrderInput) (model.Order, error)
	GetOrder(ctx context.Context, id int64) (model.Order, error)
}

type orderService struct {
	orderRepository repository.OrderRepository
}

func NewOrderService(orderRepository repository.OrderRepository) OrderService {
	return &orderService{orderRepository: orderRepository}
}

func (s *orderService) CreateOrder(ctx context.Context, input CreateOrderInput) (model.Order, error) {
	if strings.TrimSpace(input.CustomerName) == "" || input.AmountCents <= 0 {
		return model.Order{}, ErrInvalidOrderInput
	}

	return s.orderRepository.Create(ctx, repository.CreateOrderParams{
		CustomerName: strings.TrimSpace(input.CustomerName),
		AmountCents:  input.AmountCents,
	})
}

func (s *orderService) GetOrder(ctx context.Context, id int64) (model.Order, error) {
	if id <= 0 {
		return model.Order{}, ErrInvalidOrderInput
	}

	order, err := s.orderRepository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			return model.Order{}, ErrOrderNotFound
		}
		return model.Order{}, err
	}

	return order, nil
}
