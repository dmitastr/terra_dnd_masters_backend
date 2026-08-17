package masters

import (
	"context"
	"database/sql"

	"dnd_schedule/internal/domain/models"
	"github.com/sirupsen/logrus"
)

type SQLiteDS struct {
}

func NewSQLiteDS(db *sql.DB, log *logrus.Logger) IDatasource {
	return &SQLiteDS{}
}

func (S SQLiteDS) GetMasters(ctx context.Context) ([]models.Master, error) {
	// TODO implement me
	panic("implement me")
}

func (S SQLiteDS) AddMasters(ctx context.Context, masters []models.Master) ([]models.Master, error) {
	// TODO implement me
	panic("implement me")
}

func (S SQLiteDS) UpdateMasters(ctx context.Context, masters []models.Master) ([]models.Master, error) {
	// TODO implement me
	panic("implement me")
}

func (S SQLiteDS) DeleteMasters(ctx context.Context, masterIDs []int) error {
	// TODO implement me
	panic("implement me")
}
