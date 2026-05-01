package demands

import (
	"context"
	"time"

	"dnd_schedule/internal/domain/models"
)

type IDatasource interface {
	GetDemands(ctx context.Context, week time.Time) ([]models.Demand, error)
	AddDemands(ctx context.Context, demands []models.Demand) ([]models.Demand, error)
}
