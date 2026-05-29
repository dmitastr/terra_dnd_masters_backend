package demands_service

import (
	"context"
	"errors"
	"testing"
	"time"

	"dnd_schedule/internal/domain/models"
	"dnd_schedule/internal/repository/datasources/demands/mocks"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ─── Helpers ──────────────────────────────────────────────────────────────────

func newService(ds *mocks.MockIDatasource) IDemandsService {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	return NewDemandsService(ds, log)
}

func validDemand() models.Demand {
	return models.Demand{
		ID:           1,
		VkID:         42,
		FirstName:    "Ivan",
		LastName:     "Petrov",
		ForWeek:      time.Now(),
		PlayersCount: 4,
	}
}

// ─── GetDemands ───────────────────────────────────────────────────────────────

func TestGetDemands_Success(t *testing.T) {
	ds := &mocks.MockIDatasource{}
	svc := newService(ds)

	week := time.Now()
	expected := []models.Demand{validDemand()}

	ds.On("GetDemands", mock.Anything, week).Return(expected, nil)

	result, err := svc.GetDemands(context.Background(), week)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	ds.AssertExpectations(t)
}

func TestGetDemands_ReturnsEmptySlice(t *testing.T) {
	ds := &mocks.MockIDatasource{}
	svc := newService(ds)

	week := time.Now()
	ds.On("GetDemands", mock.Anything, week).Return([]models.Demand{}, nil)

	result, err := svc.GetDemands(context.Background(), week)

	assert.NoError(t, err)
	assert.Empty(t, result)
	ds.AssertExpectations(t)
}

func TestGetDemands_DatasourceError(t *testing.T) {
	ds := &mocks.MockIDatasource{}
	svc := newService(ds)

	week := time.Now()
	dsErr := errors.New("db connection failed")
	ds.On("GetDemands", mock.Anything, week).Return(nil, dsErr)

	result, err := svc.GetDemands(context.Background(), week)

	assert.Nil(t, result)
	assert.ErrorContains(t, err, "error getting masters-service")
	assert.ErrorContains(t, err, dsErr.Error())
	ds.AssertExpectations(t)
}

// ─── GetDemandsForUser ────────────────────────────────────────────────────────

func TestGetDemandsForUser_Success(t *testing.T) {
	ds := &mocks.MockIDatasource{}
	svc := newService(ds)

	week := time.Now()
	vkID := 42
	expected := []models.Demand{validDemand()}

	ds.On("GetDemandsByVkID", mock.Anything, week, vkID).Return(expected, nil)

	result, err := svc.GetDemandsForUser(context.Background(), week, vkID)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	ds.AssertExpectations(t)
}

func TestGetDemandsForUser_ReturnsEmptySlice(t *testing.T) {
	ds := &mocks.MockIDatasource{}
	svc := newService(ds)

	week := time.Now()
	vkID := 99

	ds.On("GetDemandsByVkID", mock.Anything, week, vkID).Return([]models.Demand{}, nil)

	result, err := svc.GetDemandsForUser(context.Background(), week, vkID)

	assert.NoError(t, err)
	assert.Empty(t, result)
	ds.AssertExpectations(t)
}

func TestGetDemandsForUser_DatasourceError(t *testing.T) {
	ds := &mocks.MockIDatasource{}
	svc := newService(ds)

	week := time.Now()
	vkID := 42
	dsErr := errors.New("user not found")

	ds.On("GetDemandsByVkID", mock.Anything, week, vkID).Return(nil, dsErr)

	result, err := svc.GetDemandsForUser(context.Background(), week, vkID)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, dsErr) // сервис пробрасывает ошибку as-is
	ds.AssertExpectations(t)
}

// ─── AddDemands ───────────────────────────────────────────────────────────────

func TestAddDemands_Success(t *testing.T) {
	ds := &mocks.MockIDatasource{}
	svc := newService(ds)

	input := []models.Demand{validDemand()}
	expected := []models.Demand{validDemand()}
	expected[0].ID = 10

	ds.On("AddDemands", mock.Anything, input).Return(expected, nil)

	result, err := svc.AddDemands(context.Background(), input)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	ds.AssertExpectations(t)
}

func TestAddDemands_MultipleValidDemands(t *testing.T) {
	ds := &mocks.MockIDatasource{}
	svc := newService(ds)

	d1 := validDemand()
	d2 := validDemand()
	d2.VkID = 43
	input := []models.Demand{d1, d2}

	ds.On("AddDemands", mock.Anything, input).Return(input, nil)

	result, err := svc.AddDemands(context.Background(), input)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	ds.AssertExpectations(t)
}

func TestAddDemands_InvalidDemand_ZeroVkID(t *testing.T) {
	ds := &mocks.MockIDatasource{}
	svc := newService(ds)

	bad := validDemand()
	bad.VkID = 0 // невалидный

	result, err := svc.AddDemands(context.Background(), []models.Demand{bad})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrInvalidDemands)
	ds.AssertNotCalled(t, "AddDemands") // до datasource не доходим
}

func TestAddDemands_InvalidDemand_ZeroPlayersCount(t *testing.T) {
	ds := &mocks.MockIDatasource{}
	svc := newService(ds)

	bad := validDemand()
	bad.PlayersCount = 0

	result, err := svc.AddDemands(context.Background(), []models.Demand{bad})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrInvalidDemands)
	ds.AssertNotCalled(t, "AddDemands")
}

func TestAddDemands_OneInvalidAmongMany(t *testing.T) {
	ds := &mocks.MockIDatasource{}
	svc := newService(ds)

	good := validDemand()
	bad := validDemand()
	bad.VkID = 0

	result, err := svc.AddDemands(context.Background(), []models.Demand{good, bad})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrInvalidDemands)
	ds.AssertNotCalled(t, "AddDemands")
}

func TestAddDemands_EmptySlice(t *testing.T) {
	ds := &mocks.MockIDatasource{}
	svc := newService(ds)

	ds.On("AddDemands", mock.Anything, []models.Demand{}).Return([]models.Demand{}, nil)

	result, err := svc.AddDemands(context.Background(), []models.Demand{})

	assert.NoError(t, err)
	assert.Empty(t, result)
	ds.AssertExpectations(t)
}

func TestAddDemands_DatasourceError(t *testing.T) {
	ds := &mocks.MockIDatasource{}
	svc := newService(ds)

	input := []models.Demand{validDemand()}
	dsErr := errors.New("insert failed")

	ds.On("AddDemands", mock.Anything, input).Return(nil, dsErr)

	result, err := svc.AddDemands(context.Background(), input)

	assert.Nil(t, result)
	assert.ErrorContains(t, err, "error adding demands")
	assert.ErrorContains(t, err, dsErr.Error())
	ds.AssertExpectations(t)
}

// ─── DeleteDemands ────────────────────────────────────────────────────────────

func TestDeleteDemands_Success(t *testing.T) {
	ds := &mocks.MockIDatasource{}
	svc := newService(ds)

	ids := []int{1, 2, 3}
	ds.On("DeleteDemands", mock.Anything, ids).Return(nil)

	err := svc.DeleteDemands(context.Background(), ids)

	assert.NoError(t, err)
	ds.AssertExpectations(t)
}

func TestDeleteDemands_SingleID(t *testing.T) {
	ds := &mocks.MockIDatasource{}
	svc := newService(ds)

	ids := []int{7}
	ds.On("DeleteDemands", mock.Anything, ids).Return(nil)

	err := svc.DeleteDemands(context.Background(), ids)

	assert.NoError(t, err)
	ds.AssertExpectations(t)
}

func TestDeleteDemands_EmptySlice(t *testing.T) {
	ds := &mocks.MockIDatasource{}
	svc := newService(ds)

	ids := []int{}
	ds.On("DeleteDemands", mock.Anything, ids).Return(nil)

	err := svc.DeleteDemands(context.Background(), ids)

	assert.NoError(t, err)
	ds.AssertExpectations(t)
}

func TestDeleteDemands_DatasourceError(t *testing.T) {
	ds := &mocks.MockIDatasource{}
	svc := newService(ds)

	ids := []int{1, 2}
	dsErr := errors.New("delete failed")

	ds.On("DeleteDemands", mock.Anything, ids).Return(dsErr)

	err := svc.DeleteDemands(context.Background(), ids)

	assert.ErrorContains(t, err, "error deleting demands")
	assert.ErrorContains(t, err, dsErr.Error())
	ds.AssertExpectations(t)
}
