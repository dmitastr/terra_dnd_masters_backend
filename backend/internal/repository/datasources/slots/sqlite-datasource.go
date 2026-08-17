package slots

import (
	"context"
	"database/sql"

	"dnd_schedule/internal/domain/models"
	"github.com/sirupsen/logrus"
)

type SQLiteDS struct {
	db  *sql.DB
	log *logrus.Logger
}

func NewSQLiteDS(db *sql.DB, logger *logrus.Logger) IDatasource {
	return &SQLiteDS{
		db:  db,
		log: logger,
	}
}

func (s *SQLiteDS) GetSlots(ctx context.Context, forWeek string) ([]models.Slot, error) {
	s.log.WithField("forWeek", forWeek).Debug("GetSlots from DS called")

	const query = `SELECT id, name, valid_from, valid_until, for_week_str FROM slots WHERE for_week_str = ?`

	rows, err := s.db.QueryContext(ctx, query, forWeek)
	if err != nil {
		s.log.WithError(err).Error("Error getting slots from DS")
		return nil, err
	}
	defer rows.Close()

	slots, err := scanSlots(rows)
	if err != nil {
		s.log.WithError(err).Error("Error scanning slots from DS")
		return nil, err
	}

	return slots, nil
}

func (s *SQLiteDS) GetDefaultSlots(ctx context.Context) ([]models.Slot, error) {
	s.log.Debug("GetDefaultSlots from DS called")

	const query = `SELECT id, name, valid_from, valid_until, for_week_str FROM slots_default`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		s.log.WithError(err).Error("Error getting slots from DS")
		return nil, err
	}
	defer rows.Close()

	slots, err := scanSlots(rows)
	if err != nil {
		s.log.WithError(err).Error("Error scanning slots from DS")
		return nil, err
	}

	return slots, nil
}

// scanSlots replaces pgx.CollectRows(rows, pgx.RowToStructByName[models.Slot]):
// database/sql has no equivalent struct-by-column-name scanner, so each
// column is scanned into its struct field explicitly, in the same order
// as the SELECT list.
func scanSlots(rows *sql.Rows) ([]models.Slot, error) {
	slots := make([]models.Slot, 0)

	for rows.Next() {
		var slot models.Slot

		if err := rows.Scan(
			&slot.ID,
			&slot.Name,
			&slot.ValidFrom,
			&slot.ValidUntil,
			&slot.ForWeekStr,
		); err != nil {
			return nil, err
		}

		slots = append(slots, slot)
	}

	return slots, rows.Err()
}

func (s *SQLiteDS) UpdateDefaultSlots(ctx context.Context, slots []models.Slot) ([]models.Slot, error) {
	// TODO implement me
	panic("implement me")
}

func (s *SQLiteDS) AddSlots(ctx context.Context, slots []models.Slot) ([]models.Slot, error) {
	// TODO implement me
	panic("implement me")
}

func (s *SQLiteDS) DeleteSlot(ctx context.Context, slotID int) error {
	// TODO implement me
	panic("implement me")
}

func (s *SQLiteDS) UpdateSlot(ctx context.Context, slot *models.Slot) (*models.Slot, error) {
	// TODO implement me
	panic("implement me")
}
