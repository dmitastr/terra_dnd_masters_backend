package masters_service

import (
	"context"
	"fmt"

	mastersDS "dnd_schedule/internal/datasources/masters"
	"dnd_schedule/internal/domain/models"
)

type IMastersService interface {
	GetMasters(ctx context.Context) ([]models.Master, error)
	AddMasters(ctx context.Context, masters []models.Master) error
	DeleteMasters(ctx context.Context, masters []models.Master) error
}

type MastersService struct {
	datasource mastersDS.IDatasource
}

func NewMastersService(datasource mastersDS.IDatasource) IMastersService {
	return &MastersService{datasource: datasource}
}

func (m MastersService) GetMasters(ctx context.Context) ([]models.Master, error) {
	masters, err := m.datasource.GetMasters(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting masters-service: %v", err)
	}
	return masters, err
}

func (m MastersService) AddMasters(ctx context.Context, masters []models.Master) error {
	// TODO implement me
	panic("implement me")
}

func (m MastersService) DeleteMasters(ctx context.Context, masters []models.Master) error {
	// TODO implement me
	panic("implement me")
}
