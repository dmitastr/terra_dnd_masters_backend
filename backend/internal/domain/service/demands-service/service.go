package demands_service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"dnd_schedule/internal/common/constants"
	"dnd_schedule/internal/domain/models"
	"dnd_schedule/internal/repository/datasources/demands"

	"github.com/sirupsen/logrus"
)

var (
	ErrInvalidDemands  = errors.New("invalid Demands")
	ErrInvalidVKID     = errors.New("vk id is zero")
	ErrZeroPlayerCount = errors.New("player count is zero")
	ErrDemandsNotFound = errors.New("no demands found for user")
)

type IDemandsService interface {
	GetDemands(ctx context.Context, week time.Time) ([]models.Demand, error)
	GetDemandsForUser(ctx context.Context, week time.Time, vkID int) ([]models.Demand, error)
	AddDemands(ctx context.Context, demands []models.Demand) ([]models.Demand, error)
	UpdateDemands(ctx context.Context, demands []models.Demand) ([]models.Demand, error)
	DeleteDemands(ctx context.Context, demandsIDs []int) error
	DeleteDemandsByUserIdWeek(ctx context.Context, week time.Time, vkID int) error
}

type DemandsService struct {
	datasource demands.IDatasource
	*logrus.Logger
}

func NewDemandsService(datasource demands.IDatasource, log *logrus.Logger) IDemandsService {
	return &DemandsService{datasource: datasource, Logger: log}
}

func (m *DemandsService) UpdateDemands(ctx context.Context, demands []models.Demand) ([]models.Demand, error) {
	m.WithFields(logrus.Fields{
		"count": len(demands),
	}).Debug("UpdateDemands start")

	for _, demand := range demands {
		if demand.VkID <= 0 {
			m.WithError(ErrInvalidVKID).Errorf("invalid demands: %v", demand)
			return nil, ErrInvalidVKID
		} else if demand.PlayersCount <= 0 {
			m.WithError(ErrZeroPlayerCount).Errorf("invalid demands: %v", demand)
			return nil, ErrZeroPlayerCount
		}
	}

	for i := range demands {
		demands[i].UpdatedAt = time.Now()
	}

	newDemands, err := m.datasource.UpdateDemands(ctx, demands)
	if err != nil {
		m.WithError(err).Error("error adding demands")
		return nil, fmt.Errorf("error adding demands: %v", err)
	}

	m.Debug("UpdateDemands end")
	return newDemands, nil
}

func (m *DemandsService) GetDemandsForUser(ctx context.Context, week time.Time, vkID int) ([]models.Demand, error) {
	m.WithFields(logrus.Fields{
		"week": week,
		"vkID": vkID,
	}).Debug("GetDemandsForUser start")

	weekString := week.Format(constants.Layout)
	currentDemands, err := m.datasource.GetDemandsByVkID(ctx, weekString, vkID)
	if err != nil {
		m.WithError(err).Error("error getting currentDemands")
		return nil, err
	}

	m.Debug("GetDemandsForUser end")
	return currentDemands, nil
}

func (m *DemandsService) GetDemands(ctx context.Context, week time.Time) ([]models.Demand, error) {
	m.WithFields(logrus.Fields{
		"week": week,
	}).Debug("GetDemands start")
	weekString := week.Format(constants.Layout)
	currentDemands, err := m.datasource.GetDemands(ctx, weekString)
	if err != nil {
		m.WithError(err).Error("error getting currentDemands")
		return nil, fmt.Errorf("error getting masters-service: %v", err)
	}

	m.Debug("GetDemands end")
	return currentDemands, err
}

func (m *DemandsService) AddDemands(ctx context.Context, demands []models.Demand) ([]models.Demand, error) {
	m.WithFields(logrus.Fields{
		"count": len(demands),
	}).Debug("AddDemands start")

	for i, demand := range demands {
		demands[i].FillForWeekField()
		if demand.VkID <= 0 {
			m.WithError(ErrInvalidVKID).Errorf("invalid demands: %v", demand)
			return nil, ErrInvalidVKID
		} else if demand.PlayersCount <= 0 {
			m.WithError(ErrZeroPlayerCount).Errorf("invalid demands: %v", demand)
			return nil, ErrZeroPlayerCount
		}

		now := time.Now()
		demands[i].UpdatedAt = now
		demands[i].CreatedAt = now
	}

	m.WithFields(logrus.Fields{
		"count": len(demands),
	}).Infoln("AddDemands info")

	newDemands, err := m.datasource.AddDemands(ctx, demands)
	if err != nil {
		m.WithError(err).Error("error adding demands")
		return nil, fmt.Errorf("error adding demands: %v", err)
	}

	m.Debug("AddDemands end")
	return newDemands, nil
}

func (m *DemandsService) DeleteDemands(ctx context.Context, demandsIDs []int) error {
	m.WithFields(logrus.Fields{
		"count": len(demandsIDs),
	}).Debug("DeleteDemands start")

	if err := m.datasource.DeleteDemands(ctx, demandsIDs); err != nil {
		m.WithError(err).Error("error deleting demands")
		return fmt.Errorf("error deleting demands: %v", err)
	}

	m.Debug("DeleteDemands end")
	return nil
}

func (m *DemandsService) DeleteDemandsByUserIdWeek(ctx context.Context, week time.Time, vkID int) error {
	m.WithFields(logrus.Fields{
		"count": week,
		"vkID":  vkID,
	}).Debug("DeleteDemandsByUserIdWeek start")

	demandsForUser, err := m.GetDemandsForUser(ctx, week, vkID)
	if err != nil {
		m.WithError(err).Error("error DeleteDemandsByUserIdWeek")
		return err
	}
	if len(demandsForUser) == 0 {
		return ErrDemandsNotFound
	}
	m.WithFields(logrus.Fields{
		"count": len(demandsForUser),
	}).Debug("DeleteDemands delete")
	demandIDs := make([]int, len(demandsForUser))
	for i, demand := range demandsForUser {
		demandIDs[i] = demand.ID
	}

	return m.DeleteDemands(ctx, demandIDs)
}
