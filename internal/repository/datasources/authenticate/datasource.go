package authenticate

import (
	"context"

	"dnd_schedule/internal/domain/models"
)

type IDatasource interface {
	AddUser(ctx context.Context, user *models.User) (*models.User, error)
	GetUser(ctx context.Context, user *models.User) (*models.User, error)
}
