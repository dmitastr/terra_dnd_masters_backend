package masters_service

import (
	"context"
	"errors"
	"fmt"

	"dnd_schedule/internal/domain/models"
	mastersDS "dnd_schedule/internal/repository/datasources/masters"

	"github.com/sirupsen/logrus"
)

var (
	ErrEmptyInput      = errors.New("input slice must not be empty")
	ErrEmptyIDs        = errors.New("masterIDs slice must not be empty")
	ErrInvalidMasterID = errors.New("each master must have a valid ID (> 0) for update")
	ErrEmptyMasterName = errors.New("master name must not be empty")
)

type IMastersService interface {
	GetMasters(ctx context.Context) ([]models.Master, error)
	AddMasters(ctx context.Context, masters []models.Master) error
	DeleteMasters(ctx context.Context, masterIDs []int) error
	UpdateMasters(ctx context.Context, masters []models.Master) error
}
type MastersService struct {
	datasource mastersDS.IDatasource
	log        *logrus.Logger
}

func NewMastersService(datasource mastersDS.IDatasource, log *logrus.Logger) IMastersService {
	return &MastersService{datasource: datasource, log: log}
}

// GetMasters returns all masters. Returns an empty (non-nil) slice when none exist.
func (s *MastersService) GetMasters(ctx context.Context) ([]models.Master, error) {
	s.log.Debug("GetMasters: start")

	masters, err := s.datasource.GetMasters(ctx)
	if err != nil {
		s.log.WithError(err).Error("GetMasters: datasource error")
		return nil, fmt.Errorf("GetMasters: %w", err)
	}

	// Always return an initialised slice so callers never deal with a nil range.
	if masters == nil {
		masters = []models.Master{}
	}

	s.log.WithField("count", len(masters)).Debug("GetMasters: ok")
	return masters, nil
}

// AddMasters validates and persists a batch of new masters.
// The caller does not receive the created records back; pull them via GetMasters
// if the generated IDs are needed.
func (s *MastersService) AddMasters(ctx context.Context, masters []models.Master) error {
	if len(masters) == 0 {
		s.log.Warn("AddMasters: called with empty slice")
		return ErrEmptyInput
	}

	if err := validateMastersForInsert(masters); err != nil {
		s.log.WithError(err).Warn("AddMasters: validation failed")
		return fmt.Errorf("AddMasters: %w", err)
	}

	s.log.WithField("count", len(masters)).Debug("AddMasters: start")

	if _, err := s.datasource.AddMasters(ctx, masters); err != nil {
		s.log.WithError(err).Error("AddMasters: datasource error")
		return fmt.Errorf("AddMasters: %w", err)
	}

	s.log.WithField("count", len(masters)).Debug("AddMasters: ok")
	return nil
}

// UpdateMasters validates and applies a batch of updates identified by each master's ID.
func (s *MastersService) UpdateMasters(ctx context.Context, masters []models.Master) error {
	if len(masters) == 0 {
		s.log.Warn("UpdateMasters: called with empty slice")
		return ErrEmptyInput
	}

	if err := validateMastersForUpdate(masters); err != nil {
		s.log.WithError(err).Warn("UpdateMasters: validation failed")
		return fmt.Errorf("UpdateMasters: %w", err)
	}

	s.log.WithField("count", len(masters)).Debug("UpdateMasters: start")

	if _, err := s.datasource.UpdateMasters(ctx, masters); err != nil {
		s.log.WithError(err).Error("UpdateMasters: datasource error")
		return fmt.Errorf("UpdateMasters: %w", err)
	}

	s.log.WithField("count", len(masters)).Debug("UpdateMasters: ok")
	return nil
}

// DeleteMasters removes masters by their IDs.
func (s *MastersService) DeleteMasters(ctx context.Context, masterIDs []int) error {
	if len(masterIDs) == 0 {
		s.log.Warn("DeleteMasters: called with empty slice")
		return ErrEmptyIDs
	}

	if err := validateIDs(masterIDs); err != nil {
		s.log.WithError(err).Warn("DeleteMasters: validation failed")
		return fmt.Errorf("DeleteMasters: %w", err)
	}

	s.log.WithField("ids", masterIDs).Debug("DeleteMasters: start")

	if err := s.datasource.DeleteMasters(ctx, masterIDs); err != nil {
		s.log.WithError(err).Error("DeleteMasters: datasource error")
		return fmt.Errorf("DeleteMasters: %w", err)
	}

	s.log.WithField("count", len(masterIDs)).Debug("DeleteMasters: ok")
	return nil
}

// ── validators ────────────────────────────────────────────────────────────────

// validateMastersForInsert checks fields required when creating a new master.
// ID is intentionally not validated here — it is DB-generated.
func validateMastersForInsert(masters []models.Master) error {
	for i, m := range masters {
		if m.Name == "" {
			return fmt.Errorf("%w: index %d", ErrEmptyMasterName, i)
		} else if m.ID <= 0 {
			return fmt.Errorf("%w: id %d", ErrInvalidMasterID, i)
		}
	}
	return nil
}

// validateMastersForUpdate checks that every master carries a valid ID and name.
func validateMastersForUpdate(masters []models.Master) error {
	for i, m := range masters {
		if m.ID <= 0 {
			return fmt.Errorf("%w: index %d", ErrInvalidMasterID, i)
		}
		if m.Name == "" {
			return fmt.Errorf("%w: index %d", ErrEmptyMasterName, i)
		}
	}
	return nil
}

// validateIDs ensures no zero or negative IDs were passed.
func validateIDs(ids []int) error {
	for i, id := range ids {
		if id <= 0 {
			return fmt.Errorf("%w: index %d value %d", ErrInvalidMasterID, i, id)
		}
	}
	return nil
}
