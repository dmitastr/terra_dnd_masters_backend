package datasources

import (
	"context"
	"database/sql"

	"dnd_schedule/internal/domain/models"
	"dnd_schedule/internal/repository/datasources/authenticate"
	"dnd_schedule/internal/repository/datasources/demands"
	"dnd_schedule/internal/repository/datasources/masters"
	"dnd_schedule/internal/repository/datasources/slots"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type IDatasourceProvider interface {
	SlotsDatasource
	MastersDatasource
	DemandsDatasource
	AuthDatasource
}

type datasourceProvider struct {
	SlotsDatasource
	MastersDatasource
	DemandsDatasource
	AuthDatasource
}

func NewDatasourceProvider(pool *pgxpool.Pool, log *logrus.Logger) IDatasourceProvider {
	return &datasourceProvider{
		SlotsDatasource:   slots.NewSlotsDS(pool, log),
		MastersDatasource: masters.NewMastersDS(pool, log),
		DemandsDatasource: demands.NewDemandsDS(pool, log),
		AuthDatasource:    authenticate.NewAuthDatasource(pool),
	}

}

func NewSQLiteDatasourceProvider(db *sql.DB, log *logrus.Logger) IDatasourceProvider {
	return &datasourceProvider{
		SlotsDatasource:   slots.NewSQLiteDS(db, log),
		MastersDatasource: masters.NewSQLiteDS(db, log),
		DemandsDatasource: demands.NewSQLiteDemandsDS(db, log),
		AuthDatasource:    authenticate.NewSQLiteDS(db, log),
	}

}

type SlotsDatasource interface {
	GetSlots(ctx context.Context, forWeek string) ([]models.Slot, error)
	GetDefaultSlots(ctx context.Context) ([]models.Slot, error)
	UpdateDefaultSlots(ctx context.Context, slots []models.Slot) ([]models.Slot, error)
	AddSlots(ctx context.Context, slots []models.Slot) ([]models.Slot, error)
	DeleteSlot(ctx context.Context, slotID int) error
	UpdateSlot(ctx context.Context, slot *models.Slot) (*models.Slot, error)
}

type MastersDatasource interface {
	GetMasters(ctx context.Context) ([]models.Master, error)
	AddMasters(ctx context.Context, masters []models.Master) ([]models.Master, error)
	UpdateMasters(ctx context.Context, masters []models.Master) ([]models.Master, error)
	DeleteMasters(ctx context.Context, masterIDs []int) error
}

type DemandsDatasource interface {
	GetDemands(ctx context.Context, week string) ([]models.Demand, error)
	GetDemandsByVkID(ctx context.Context, week string, vkID int) ([]models.Demand, error)
	AddDemands(ctx context.Context, demands []models.Demand) ([]models.Demand, error)
	UpdateDemands(ctx context.Context, demands []models.Demand) ([]models.Demand, error)
	DeleteDemands(ctx context.Context, demandIDs []int) error
}

type AuthDatasource interface {
	AddUser(ctx context.Context, user *models.User) (*models.User, error)
	GetUser(ctx context.Context, username string) (*models.User, error)
}
