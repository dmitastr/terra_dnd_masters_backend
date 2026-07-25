package slots_service

import (
	"context"
	"fmt"
	"time"

	"dnd_schedule/internal/common/constants"
	"dnd_schedule/internal/domain/models"
	"dnd_schedule/internal/repository/datasources/slots"

	"github.com/sirupsen/logrus"
)

type ISlotsService interface {
	GetSlots(ctx context.Context, from string, to string) ([]models.Slot, error)
	AddSlots(ctx context.Context, slots []models.Slot) error
	DeleteSlots(ctx context.Context, slotIDs []int) error
	UpdateSlots(ctx context.Context, slots []models.Slot) error
}

type SlotsService struct {
	datasource slots.IDatasource
	log        *logrus.Logger
}

func (m SlotsService) UpdateSlots(ctx context.Context, slots []models.Slot) error {
	// TODO implement me
	panic("implement me")
}

func NewSlotsService(datasource slots.IDatasource, logger *logrus.Logger) ISlotsService {
	return &SlotsService{datasource: datasource, log: logger}
}

func (m SlotsService) GetSlots(ctx context.Context, from string, to string) ([]models.Slot, error) {
	m.log.WithFields(logrus.Fields{
		"from": from,
		"to":   to,
	}).Debug("GetSlots service call")

	fromDT, err := time.Parse(constants.Layout, from)
	if err != nil {
		m.log.WithError(err).Error("cant't parse fromDT")
		return nil, err
	}
	toDT, err := time.Parse(constants.Layout, to)
	if err != nil {
		m.log.WithError(err).Error("cant't parse toDT")
		return nil, err
	}

	currentSlots, err := m.datasource.GetSlots(ctx, fromDT, toDT)
	if err != nil {
		m.log.WithError(err).Error("GetSlots service call failed")
		return nil, fmt.Errorf("error getting slots-service: %v", err)
	}

	if len(currentSlots) == 0 {
		m.log.Debug("No slots found, fallback to default")
		currentSlots, err = m.datasource.GetDefaultSlots(ctx)
	}
	if err != nil {
		m.log.WithError(err).Error("GetSlots service call failed")
		return nil, fmt.Errorf("error getting slots-service: %v", err)
	}

	m.log.WithFields(logrus.Fields{
		"count": len(currentSlots),
	}).Debug("GetSlots service call success")
	return currentSlots, err
}

func (m SlotsService) AddSlots(ctx context.Context, slots []models.Slot) error {
	// TODO implement me
	panic("implement me")
}

func (m SlotsService) DeleteSlots(ctx context.Context, slotIDs []int) error {
	// TODO implement me
	panic("implement me")
}
