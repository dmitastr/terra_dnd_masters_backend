package slots

import (
	"context"
	"fmt"
	"time"

	"dnd_schedule/internal/domain/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type IDatasource interface {
	GetSlots(ctx context.Context, from, to time.Time) ([]models.Slot, error)
	GetDefaultSlots(ctx context.Context) ([]models.Slot, error)
	UpdateDefaultSlots(ctx context.Context, slots []models.Slot) ([]models.Slot, error)
	AddSlots(ctx context.Context, slots []models.Slot) ([]models.Slot, error)
	DeleteSlot(ctx context.Context, slotID int) error
	UpdateSlot(ctx context.Context, slot *models.Slot) (*models.Slot, error)
}

type SlotsDS struct {
	pool *pgxpool.Pool
	log  *logrus.Logger
}

func NewSlotsDS(pool *pgxpool.Pool, log *logrus.Logger) *SlotsDS {
	return &SlotsDS{pool: pool, log: log}
}

func (s *SlotsDS) UpdateDefaultSlots(ctx context.Context, slots []models.Slot) ([]models.Slot, error) {
	s.log.WithField("slots", slots).Debug("UpdateDefaultSlots from DS called")
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		s.log.WithError(err).Error("error starting transaction")
		return nil, err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM slots_default`); err != nil {
		_ = tx.Rollback(ctx)
		s.log.WithError(err).Error("error deleting default slot")
		return nil, err
	}

	batch := &pgx.Batch{}

	for _, slot := range slots {
		batch.Queue(
			`INSERT INTO slots_default (id, name) VALUES($1, $2)`,
			slot.ID,
			slot.Name,
		)
	}
	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	for range slots {
		_, err := br.Exec()
		if err != nil {
			s.log.WithError(err).Error("Error adding default slots to DS")
			_ = tx.Rollback(ctx)

			return nil, fmt.Errorf("insert slots: %w", err)
		}
	}
	return slots, nil
}

func (s *SlotsDS) GetSlots(ctx context.Context, from, to time.Time) ([]models.Slot, error) {
	s.log.WithField("from", from).WithField("to", to).Debug("GetSlots from DS called")
	query := `SELECT id, name FROM slots WHERE valid_from BETWEEN $1 AND $2 AND valid_to BETWEEN $1 AND $2;`
	rows, err := s.pool.Query(ctx, query, from, to)
	if err != nil {
		s.log.WithError(err).Error("Error getting slots from DS")
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[models.Slot])
}

func (s *SlotsDS) GetDefaultSlots(ctx context.Context) ([]models.Slot, error) {
	s.log.Debug("GetDefaultSlots from DS called")
	query := `SELECT id, name FROM slots_default;`
	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		s.log.WithError(err).Error("Error getting slots from DS")
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[models.Slot])
}

func (s *SlotsDS) AddSlots(ctx context.Context, slots []models.Slot) ([]models.Slot, error) {
	s.log.WithField("slots", slots).Debug("AddSlots from DS called")
	batch := &pgx.Batch{}

	for _, slot := range slots {
		batch.Queue(
			`INSERT INTO slots (id, name, valid_from, valid_until) VALUES($1, $2, $3, $4)`,
			slot.ID,
			slot.Name,
			slot.ValidFrom,
			slot.ValidUntil,
		)
	}
	br := s.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range slots {
		_, err := br.Exec()
		if err != nil {
			s.log.WithError(err).Error("Error adding slots to DS")

			return nil, fmt.Errorf("insert user: %w", err)
		}
	}
	return slots, nil

}

func (s *SlotsDS) DeleteSlot(ctx context.Context, slotID int) error {
	// TODO implement me
	panic("implement me")
}

func (s *SlotsDS) UpdateSlot(ctx context.Context, slot *models.Slot) (*models.Slot, error) {
	// TODO implement me
	panic("implement me")
}
