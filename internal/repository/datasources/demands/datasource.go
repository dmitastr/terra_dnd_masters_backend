package demands

import (
	"context"
	"fmt"
	"time"

	"dnd_schedule/internal/domain/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type IDatasource interface {
	GetDemands(ctx context.Context, week time.Time) ([]models.Demand, error)
	GetDemandsByVkID(ctx context.Context, week time.Time, vkID int) ([]models.Demand, error)
	AddDemands(ctx context.Context, demands []models.Demand) ([]models.Demand, error)
	UpdateDemand(ctx context.Context, demand *models.Demand) (*models.Demand, error)
	DeleteDemand(ctx context.Context, demand *models.Demand) error
}

type DemandsDS struct {
	pool *pgxpool.Pool
	log  *logrus.Logger
}

func NewDemandsDS(pool *pgxpool.Pool, log *logrus.Logger) IDatasource {
	return &DemandsDS{pool: pool, log: log}
}

func (d DemandsDS) GetDemandsByVkID(ctx context.Context, week time.Time, vkID int) ([]models.Demand, error) {
	d.log.WithField("week", week).Info("getting demands by vkID")
	query := `SELECT id, vk_id, first_name, last_name, players_count, for_week, slots
	FROM demands WHERE for_week = $1 AND vk_id = $2`

	rows, err := d.pool.Query(ctx, query, week, vkID)
	if err != nil {
		d.log.WithError(err).Error("error getting demands")
		return nil, err
	}
	defer rows.Close()

	demands := make([]models.Demand, 0)
	for rows.Next() {
		demand := models.Demand{}
		err := rows.Scan(
			&demand.ID,
			&demand.VkID,
			&demand.FirstName,
			&demand.LastName,
			&demand.PlayersCount,
			&demand.ForWeek,
			&demand.Slots,
		)
		if err != nil {
			d.log.WithError(err).Error("error getting demands")
			return nil, err
		}
		demands = append(demands, demand)
	}
	return demands, nil
}

func (d DemandsDS) GetDemands(ctx context.Context, week time.Time) ([]models.Demand, error) {
	d.log.WithField("week", week).Info("getting demands")
	query := `SELECT id, vk_id, first_name, last_name, players_count, for_week, slots
	FROM demands WHERE for_week = $1`

	rows, err := d.pool.Query(ctx, query, week)
	if err != nil {
		d.log.WithError(err).Error("error getting demands")
		return nil, err
	}
	defer rows.Close()

	demands := make([]models.Demand, 0)
	for rows.Next() {
		demand := models.Demand{}
		err := rows.Scan(
			&demand.ID,
			&demand.VkID,
			&demand.FirstName,
			&demand.LastName,
			&demand.PlayersCount,
			&demand.ForWeek,
			&demand.Slots,
		)
		if err != nil {
			d.log.WithError(err).Error("error getting demands")
			return nil, err
		}
		demands = append(demands, demand)
	}
	return demands, nil
}

func (d DemandsDS) AddDemands(ctx context.Context, demands []models.Demand) ([]models.Demand, error) {
	d.log.WithField("slots", demands).Debug("AddDemands from DS called")
	batch := &pgx.Batch{}

	for _, demand := range demands {
		batch.Queue(
			`INSERT INTO demands (vk_id, first_name, last_name, for_week, players_count, slots) 
                   VALUES($1, $2, $3, $4, $5, $6) RETURN id`,
			demand.VkID,
			demand.FirstName,
			demand.LastName,
			demand.ForWeek,
			demand.PlayersCount,
			demand.Slots,
		)
	}
	br := d.pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := range demands {
		if err := br.QueryRow().Scan(&demands[i].ID); err != nil {
			d.log.WithError(err).Error("Error adding demands to DS")

			return nil, fmt.Errorf("insert demand: %w", err)
		}
	}
	return demands, nil
}

func (d DemandsDS) UpdateDemand(ctx context.Context, demand *models.Demand) (*models.Demand, error) {
	d.log.WithField("demand", demand).Info("updating demand")
	query := `UPDATE demands SET slots = $1 WHERE id = $2`
	_, err := d.pool.Exec(ctx, query, demand.Slots, demand.ID)
	if err != nil {
		d.log.WithError(err).Error("error updating demand")
		return nil, err
	}
	return demand, nil
}

func (d DemandsDS) DeleteDemand(ctx context.Context, demand *models.Demand) error {
	// TODO implement me
	panic("implement me")
}
