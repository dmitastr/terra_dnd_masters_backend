package demands_service

import (
	"context"
	"fmt"
	"time"

	"dnd_schedule/internal/domain/models"
	"dnd_schedule/internal/repository/datasources/demands"
)

type IDemandsService interface {
	GetDemands(ctx context.Context, week time.Time) ([]models.Demand, error)
	AddDemands(ctx context.Context, demands []models.Demand) ([]models.Demand, error)
	DeleteDemands(ctx context.Context, masters []models.Demand) error
}

type DemandsService struct {
	datasource demands.IDatasource
}

func NewDemandsService(datasource demands.IDatasource) IDemandsService {
	return &DemandsService{datasource: datasource}
}

func (m DemandsService) GetDemands(ctx context.Context, week time.Time) ([]models.Demand, error) {
	currentDemands, err := m.datasource.GetDemands(ctx, week)
	if err != nil {
		return nil, fmt.Errorf("error getting masters-service: %v", err)
	}
	return currentDemands, err
}

func (m DemandsService) AddDemands(ctx context.Context, demands []models.Demand) ([]models.Demand, error) {
	newDemands, err := m.datasource.AddDemands(ctx, demands)
	if err != nil {
		return nil, fmt.Errorf("error adding demands: %v", err)
	}
	return newDemands, nil
}

func (m DemandsService) DeleteDemands(ctx context.Context, masters []models.Demand) error {
	// TODO implement me
	panic("implement me")
}
