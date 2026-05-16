package service

import "order-system/internal/repository"

type HealthService interface {
	GetStatus() string
}

type healthService struct {
	healthRepository repository.HealthRepository
}

func NewHealthService(healthRepository repository.HealthRepository) HealthService {
	return healthService{healthRepository: healthRepository}
}

func (s healthService) GetStatus() string {
	return s.healthRepository.Status()
}
