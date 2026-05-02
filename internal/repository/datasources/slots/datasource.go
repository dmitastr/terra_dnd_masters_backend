package slots

import (
	"context"

	"dnd_schedule/internal/domain/models"
)

type IDatasource interface {
	GetSlots(ctx context.Context) ([]models.Slot, error)
}
