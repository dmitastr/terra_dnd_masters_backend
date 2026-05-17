package datasource

import (
	"context"
	"time"

	"dnd_schedule/internal/domain/models"
	testData "dnd_schedule/internal/testing/data"
)

type DummyDatasource struct {
	demandsDS *testData.DemandsDS
}

func NewDatasource() *DummyDatasource {
	return &DummyDatasource{demandsDS: testData.NewDemandsDS()}
}

func (d DummyDatasource) AddUser(ctx context.Context, user *models.User) (*models.User, error) {
	return user, nil
}

func (d DummyDatasource) GetUser(ctx context.Context, user *models.User) (*models.User, error) {
	return user, nil
}

func (d DummyDatasource) GetMasters(ctx context.Context) ([]models.Master, error) {
	return testData.Masters, nil
}

func (d DummyDatasource) GetSlots(ctx context.Context) ([]models.Slot, error) {
	return testData.Slots, nil
}

func (d DummyDatasource) GetDemands(ctx context.Context, week time.Time) ([]models.Demand, error) {
	return d.demandsDS.GetDemands(), nil
}
func (d DummyDatasource) AddDemands(ctx context.Context, demands []models.Demand) ([]models.Demand, error) {
	d.demandsDS.AddDemands(demands)
	return demands, nil
}
