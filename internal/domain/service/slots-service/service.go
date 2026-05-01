package slots_service

import (
	"context"
	"fmt"

	"dnd_schedule/internal/datasources/slots"
	"dnd_schedule/internal/domain/models"
)

type ISlotsService interface {
	GetSlots(ctx context.Context) ([]models.Slot, error)
	AddSlots(ctx context.Context, masters []models.Slot) error
	DeleteSlots(ctx context.Context, masters []models.Slot) error
}

type SlotsService struct {
	datasource slots.IDatasource
}

func NewSlotsService(datasource slots.IDatasource) ISlotsService {
	return &SlotsService{datasource: datasource}
}

func (m SlotsService) GetSlots(ctx context.Context) ([]models.Slot, error) {
	currentSlots, err := m.datasource.GetSlots(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting slots-service: %v", err)
	}
	return currentSlots, err
}

func (m SlotsService) AddSlots(ctx context.Context, newSlots []models.Slot) error {
	// TODO implement me
	panic("implement me")
}

func (m SlotsService) DeleteSlots(ctx context.Context, removingSlots []models.Slot) error {
	// TODO implement me
	panic("implement me")
}
