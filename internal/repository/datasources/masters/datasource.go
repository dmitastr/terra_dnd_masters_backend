package masters

import (
	"context"

	"dnd_schedule/internal/domain/models"
)

type IDatasource interface {
	GetMasters(ctx context.Context) ([]models.Master, error)
}
