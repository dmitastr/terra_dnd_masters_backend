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
	UpdateDemands(ctx context.Context, demands []models.Demand) ([]models.Demand, error)
	DeleteDemands(ctx context.Context, demandIDs []int) error
}

type DemandsDS struct {
	pool *pgxpool.Pool
	log  *logrus.Logger
}

func NewDemandsDS(pool *pgxpool.Pool, log *logrus.Logger) IDatasource {
	return &DemandsDS{pool: pool, log: log}
}

func (d DemandsDS) DeleteDemands(ctx context.Context, demandIDs []int) error {
	if len(demandIDs) == 0 {
		return nil
	}

	const query = `DELETE FROM demands WHERE id = ANY($1::int[])`

	tag, err := d.pool.Exec(ctx, query, demandIDs)
	if err != nil {
		d.log.WithError(err).WithField("ids", demandIDs).Error("DeleteDemands: exec failed")
		return fmt.Errorf("DeleteDemands: %w", err)
	}

	d.log.WithFields(logrus.Fields{
		"requested": len(demandIDs),
		"deleted":   tag.RowsAffected(),
	}).Debug("DeleteDemands: ok")

	return nil
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

func (d DemandsDS) UpdateDemands(ctx context.Context, demands []models.Demand) ([]models.Demand, error) {
	if len(demands) == 0 {
		return nil, nil
	}

	const query = `
        UPDATE demands AS t
        SET
            slots       = v.slots,
            players_count = v.players_count,
            updated_at = v.updated_at
        FROM (
            SELECT
                UNNEST($1::int[])         AS vk_id,
                UNNEST($2::text[])        AS for_week,
                UNNEST($3::timestamptz[]) AS updated_at,
                UNNEST($5::jsonb[]) AS slots
        ) AS v
        WHERE t.vk_id = v.vk_id AND v.for_week = t.for_week
        RETURNING t.id, t.created_at, t.updated_at, t.vk_id`

	ids := make([]int, len(demands))
	vkIDs := make([]int, len(demands))
	forWeeks := make([]interface{}, len(demands))
	slots := make([]interface{}, len(demands))
	updatedAts := make([]interface{}, len(demands))

	for i, m := range demands {
		ids[i] = m.ID
		vkIDs[i] = m.VkID
		updatedAts[i] = m.UpdatedAt
		forWeeks[i] = m.ForWeek
		slots[i] = m.Slots
	}

	rows, err := d.pool.Query(ctx, query, ids, vkIDs, updatedAts)
	if err != nil {
		d.log.WithError(err).WithField("count", len(demands)).Error("UpdateDemands: query failed")
		return nil, fmt.Errorf("UpdateDemands: %w", err)
	}
	defer rows.Close()

	updated, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Demand])
	if err != nil {
		d.log.WithError(err).Error("UpdateDemands: scanning rows failed")
		return nil, fmt.Errorf("UpdateDemands: %w", err)
	}

	if len(updated) != len(demands) {
		d.log.WithFields(logrus.Fields{
			"requested": len(demands),
			"updated":   len(updated),
		}).Warn("UpdateDemands: some IDs were not found")
	}

	d.log.WithField("count", len(updated)).Debug("UpdateDemands: ok")
	return updated, nil
}
