package masters

import (
	"context"
	"fmt"

	"dnd_schedule/internal/domain/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type IDatasource interface {
	GetMasters(ctx context.Context) ([]models.Master, error)
	AddMasters(ctx context.Context, masters []models.Master) ([]models.Master, error)
	UpdateMasters(ctx context.Context, masters []models.Master) ([]models.Master, error)
	DeleteMasters(ctx context.Context, masterIDs []int) error
}

type MastersDS struct {
	pool *pgxpool.Pool
	log  *logrus.Logger
}

func NewMastersDS(pool *pgxpool.Pool, log *logrus.Logger) *MastersDS {
	return &MastersDS{pool: pool, log: log}
}

// GetMasters fetches all masters from the database.
func (ds *MastersDS) GetMasters(ctx context.Context) ([]models.Master, error) {
	const query = `
        SELECT id, name, created_at, updated_at, vk_id
        FROM masters
        ORDER BY id`

	rows, err := ds.pool.Query(ctx, query)
	if err != nil {
		ds.log.WithError(err).Error("GetMasters: query failed")
		return nil, fmt.Errorf("GetMasters: %w", err)
	}
	defer rows.Close()

	masters, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Master])
	if err != nil {
		ds.log.WithError(err).Error("GetMasters: scanning rows failed")
		return nil, fmt.Errorf("GetMasters: %w", err)
	}

	ds.log.WithField("count", len(masters)).Debug("GetMasters: ok")
	return masters, nil
}

// AddMasters bulk-inserts masters and returns them with their generated IDs.
// Uses unnest to send a single round-trip regardless of slice size.
func (ds *MastersDS) AddMasters(ctx context.Context, masters []models.Master) ([]models.Master, error) {
	if len(masters) == 0 {
		return nil, nil
	}

	const query = `
        INSERT INTO masters (name, created_at, updated_at, vk_id)
        SELECT
            UNNEST($1::text[]),
            UNNEST($2::timestamptz[]),
            UNNEST($3::timestamptz[])
			UNNEST($4::int[])
        RETURNING id, name, created_at, updated_at, vk_id`

	names := make([]string, len(masters))
	createdAts := make([]interface{}, len(masters))
	updatedAts := make([]interface{}, len(masters))
	vkIDs := make([]interface{}, len(masters))

	for i, m := range masters {
		names[i] = m.Name
		createdAts[i] = m.CreatedAt
		updatedAts[i] = m.UpdatedAt
		vkIDs[i] = m.VkID
	}

	rows, err := ds.pool.Query(ctx, query, names, createdAts, updatedAts)
	if err != nil {
		ds.log.WithError(err).WithField("count", len(masters)).Error("AddMasters: query failed")
		return nil, fmt.Errorf("AddMasters: %w", err)
	}
	defer rows.Close()

	inserted, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Master])
	if err != nil {
		ds.log.WithError(err).Error("AddMasters: scanning rows failed")
		return nil, fmt.Errorf("AddMasters: %w", err)
	}

	ds.log.WithField("count", len(inserted)).Debug("AddMasters: ok")
	return inserted, nil
}

// UpdateMasters bulk-updates masters by ID.
// Uses a temporary VALUES list joined against the target table so all rows
// are updated in one statement — no N+1 round-trips.
func (ds *MastersDS) UpdateMasters(ctx context.Context, masters []models.Master) ([]models.Master, error) {
	if len(masters) == 0 {
		return nil, nil
	}

	const query = `
        UPDATE masters AS t
        SET
            name       = v.name,
            updated_at = v.updated_at
        FROM (
            SELECT
                UNNEST($1::int[])         AS id,
                UNNEST($2::text[])        AS name,
                UNNEST($3::timestamptz[]) AS updated_at
        ) AS v
        WHERE t.id = v.id
        RETURNING t.id, t.name, t.created_at, t.updated_at, t.vk_id`

	ids := make([]int, len(masters))
	names := make([]string, len(masters))
	updatedAts := make([]interface{}, len(masters))

	for i, m := range masters {
		ids[i] = m.ID
		names[i] = m.Name
		updatedAts[i] = m.UpdatedAt
	}

	rows, err := ds.pool.Query(ctx, query, ids, names, updatedAts)
	if err != nil {
		ds.log.WithError(err).WithField("count", len(masters)).Error("UpdateMasters: query failed")
		return nil, fmt.Errorf("UpdateMasters: %w", err)
	}
	defer rows.Close()

	updated, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Master])
	if err != nil {
		ds.log.WithError(err).Error("UpdateMasters: scanning rows failed")
		return nil, fmt.Errorf("UpdateMasters: %w", err)
	}

	if len(updated) != len(masters) {
		ds.log.WithFields(logrus.Fields{
			"requested": len(masters),
			"updated":   len(updated),
		}).Warn("UpdateMasters: some IDs were not found")
	}

	ds.log.WithField("count", len(updated)).Debug("UpdateMasters: ok")
	return updated, nil
}

// DeleteMasters removes masters by their IDs in a single statement.
func (ds *MastersDS) DeleteMasters(ctx context.Context, masterIDs []int) error {
	if len(masterIDs) == 0 {
		return nil
	}

	const query = `DELETE FROM masters WHERE id = ANY($1::int[])`

	tag, err := ds.pool.Exec(ctx, query, masterIDs)
	if err != nil {
		ds.log.WithError(err).WithField("ids", masterIDs).Error("DeleteMasters: exec failed")
		return fmt.Errorf("DeleteMasters: %w", err)
	}

	ds.log.WithFields(logrus.Fields{
		"requested": len(masterIDs),
		"deleted":   tag.RowsAffected(),
	}).Debug("DeleteMasters: ok")

	return nil
}
