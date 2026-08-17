package authenticate

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

func (S SQLiteDS) AddUser(ctx context.Context, user *models.User) (*models.User, error) {
	// TODO implement me
	panic("implement me")
}

func (S SQLiteDS) GetUser(ctx context.Context, username string) (*models.User, error) {
	// TODO implement me
	panic("implement me")
}
