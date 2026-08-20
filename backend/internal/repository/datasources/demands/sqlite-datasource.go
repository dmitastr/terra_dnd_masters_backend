package demands

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"dnd_schedule/internal/domain/models"

	"github.com/sirupsen/logrus"
	_ "modernc.org/sqlite" // pure-Go sqlite driver, registers as "sqlite"
)

// NOTE on porting from Postgres to SQLite — assumptions made below, please
// adjust if they don't match your actual schema/models:
//
//  1. SQLite has no array types, so ANY($1::int[]) and UNNEST(...) are gone.
//     DeleteDemands now builds a `WHERE id IN (?, ?, ...)` clause, and
//     UpdateDemands loops over the input rows in a single transaction
//     instead of doing one set-based UPDATE. For a handful of rows per
//     call this is effectively the same cost, and it avoids surprises from
//     driver-specific array encoding.
//
//  2. AddDemands keeps the same upsert shape (INSERT ... ON CONFLICT ...
//     DO UPDATE ... RETURNING id) — SQLite (3.24+ for upsert, 3.35+ for
//     RETURNING) supports both. This requires a UNIQUE constraint/index on
//     (vk_id, for_week_str) in your migrations, same as Postgres needed.
//
//  3. The original UpdateDemands had a bug: its query referenced
//     v.players_count from the UNNEST subquery, but players_count was
//     never included in the UNNEST args (only 3 args were passed for 4
//     placeholders). Rather than port the bug, UpdateDemands here matches
//     and updates by primary key (id), which is unambiguous and doesn't
//     depend on (vk_id, for_week) being unique.
//
//  4. models.Demand's CreatedAt/UpdatedAt are assumed to be time.Time, and
//     Slots is assumed to already implement database/sql's driver.Valuer /
//     sql.Scanner (as it must have, to be written directly to a Postgres
//     jsonb column) — those interfaces are driver-agnostic, so the same
//     type should Value()/Scan() fine against a SQLite TEXT/BLOB column.
//     If Slots is actually a pgtype-specific wrapper (e.g. pgtype.JSONB),
//     it won't work as-is and needs to become []byte or a plain type with
//     its own Value()/Scan() implemented in terms of encoding/json.
//
//  5. Columns storing time.Time (created_at, updated_at) should be declared
//     DATETIME/TIMESTAMP in the migration; modernc.org/sqlite scans those
//     into time.Time via RFC3339 text under the hood.

type SQLiteDemandsDS struct {
	db  *sql.DB
	log *logrus.Logger
}

func NewSQLiteDemandsDS(db *sql.DB, log *logrus.Logger) IDatasource {
	return &SQLiteDemandsDS{db: db, log: log}
}

func (d *SQLiteDemandsDS) DeleteDemands(ctx context.Context, demandIDs []int) error {
	if len(demandIDs) == 0 {
		return nil
	}

	placeholders := make([]string, len(demandIDs))
	args := make([]any, len(demandIDs))
	for i, id := range demandIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`DELETE FROM demands WHERE id IN (%s)`, strings.Join(placeholders, ", "))

	res, err := d.db.ExecContext(ctx, query, args...)
	if err != nil {
		d.log.WithError(err).WithField("ids", demandIDs).Error("DeleteDemands: exec failed")
		return fmt.Errorf("DeleteDemands: %w", err)
	}

	deleted, err := res.RowsAffected()
	if err != nil {
		d.log.WithError(err).Warn("DeleteDemands: could not read rows affected")
	}

	d.log.WithFields(logrus.Fields{
		"requested": len(demandIDs),
		"deleted":   deleted,
	}).Debug("DeleteDemands: ok")

	return nil
}

func (d *SQLiteDemandsDS) GetDemandsByVkID(ctx context.Context, week string, vkID int) ([]models.Demand, error) {
	d.log.WithField("week", week).WithField("vkID", vkID).Info("getting demands by vkID")

	const query = `
	SELECT id, vk_id, first_name, last_name, players_count, for_week, slots, created_at, updated_at, vk_username, COALESCE(comment, '') as comment
	FROM demands WHERE for_week_str = ? AND vk_id = ?
    ORDER BY updated_at DESC`

	rows, err := d.db.QueryContext(ctx, query, week, vkID)
	if err != nil {
		d.log.WithError(err).Error("error getting demands")
		return nil, err
	}
	defer rows.Close()

	demands, err := scanDemands(rows)
	if err != nil {
		d.log.WithError(err).Error("error iterating demands")
		return nil, err
	}

	return demands, nil
}

func (d *SQLiteDemandsDS) GetDemands(ctx context.Context, week string) ([]models.Demand, error) {
	d.log.WithField("week", week).Info("getting demands")

	const query = `SELECT
    id, vk_id, first_name, last_name, players_count, for_week, slots, created_at, updated_at, vk_username, COALESCE(comment, '') as comment
    FROM demands WHERE for_week_str = ?`

	// database/sql's Query never returns sql.ErrNoRows for an empty result
	// set (unlike QueryRow) — that Postgres-specific special case from the
	// pgx version isn't needed here.
	rows, err := d.db.QueryContext(ctx, query, week)
	if err != nil {
		d.log.WithError(err).Error("error getting demands")
		return nil, err
	}
	defer rows.Close()

	demands, err := scanDemands(rows)
	if err != nil {
		d.log.WithError(err).Error("error iterating demands")
		return nil, err
	}

	return demands, nil
}

func scanDemands(rows *sql.Rows) ([]models.Demand, error) {
	demands := make([]models.Demand, 0)

	for rows.Next() {
		var demand models.Demand
		var slotsRaw []byte

		if err := rows.Scan(
			&demand.ID,
			&demand.VkID,
			&demand.FirstName,
			&demand.LastName,
			&demand.PlayersCount,
			&demand.ForWeek,
			&slotsRaw,
			&demand.CreatedAt,
			&demand.UpdatedAt,
			&demand.VkUsername,
			&demand.Comment,
		); err != nil {
			return nil, err
		}
		if len(slotsRaw) > 0 {
			if err := json.Unmarshal(slotsRaw, &demand.Slots); err != nil {
				return nil, fmt.Errorf("decode slots: %w", err)
			}
		}

		demands = append(demands, demand)
	}

	return demands, rows.Err()
}

func (d *SQLiteDemandsDS) AddDemands(ctx context.Context, demands []models.Demand) ([]models.Demand, error) {
	d.log.WithField("slots", demands).Debug("AddDemands from DS called")

	if len(demands) == 0 {
		return nil, nil
	}

	const query = `
		INSERT INTO demands (vk_id, first_name, last_name, for_week, players_count, slots, created_at, updated_at, vk_username, for_week_str, comment)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(vk_id, for_week_str)
		DO UPDATE SET
			first_name    = excluded.first_name,
			last_name     = excluded.last_name,
			players_count = excluded.players_count,
			slots         = excluded.slots,
			updated_at    = excluded.updated_at,
			vk_username   = excluded.vk_username,
			comment       = excluded.comment
		RETURNING id`

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		d.log.WithError(err).Error("AddDemands: begin tx failed")
		return nil, fmt.Errorf("AddDemands: begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // no-op once committed

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		d.log.WithError(err).Error("AddDemands: prepare failed")
		return nil, fmt.Errorf("AddDemands: prepare: %w", err)
	}
	defer stmt.Close()

	for i, demand := range demands {
		slotsRaw, err := json.Marshal(demand.Slots)
		if err != nil {
			return nil, fmt.Errorf("encode slots: %w", err)
		}
		err = stmt.QueryRowContext(ctx,
			demand.VkID,
			demand.FirstName,
			demand.LastName,
			demand.ForWeek,
			demand.PlayersCount,
			slotsRaw,
			demand.CreatedAt,
			demand.UpdatedAt,
			demand.VkUsername,
			demand.ForWeekStr,
			demand.Comment,
		).Scan(&demands[i].ID)
		if err != nil {
			d.log.WithError(err).Error("Error adding demands to DS")
			return nil, fmt.Errorf("insert demand: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		d.log.WithError(err).Error("AddDemands: commit failed")
		return nil, fmt.Errorf("AddDemands: commit: %w", err)
	}

	return demands, nil
}

func (d *SQLiteDemandsDS) UpdateDemands(ctx context.Context, demands []models.Demand) ([]models.Demand, error) {
	if len(demands) == 0 {
		return nil, nil
	}

	// See the package-level note: this matches/updates by primary key (id)
	// rather than porting the original's broken vk_id+for_week UNNEST update.
	const query = `
		UPDATE demands
		SET slots = ?, players_count = ?, updated_at = ?
		WHERE id = ?
		RETURNING id, created_at, updated_at, vk_id`

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		d.log.WithError(err).Error("UpdateDemands: begin tx failed")
		return nil, fmt.Errorf("UpdateDemands: begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // no-op once committed

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		d.log.WithError(err).Error("UpdateDemands: prepare failed")
		return nil, fmt.Errorf("UpdateDemands: prepare: %w", err)
	}
	defer stmt.Close()

	updated := make([]models.Demand, 0, len(demands))
	for _, m := range demands {
		var (
			id        int
			createdAt time.Time
			updatedAt time.Time
			vkID      int
		)

		err := stmt.QueryRowContext(ctx, m.Slots, m.PlayersCount, m.UpdatedAt, m.ID).
			Scan(&id, &createdAt, &updatedAt, &vkID)
		if err != nil {
			if err == sql.ErrNoRows {
				d.log.WithField("id", m.ID).Warn("UpdateDemands: id not found")
				continue
			}
			d.log.WithError(err).WithField("id", m.ID).Error("UpdateDemands: query failed")
			return nil, fmt.Errorf("UpdateDemands: %w", err)
		}

		m.ID = id
		m.CreatedAt = createdAt
		m.UpdatedAt = updatedAt
		m.VkID = vkID
		updated = append(updated, m)
	}

	if err := tx.Commit(); err != nil {
		d.log.WithError(err).Error("UpdateDemands: commit failed")
		return nil, fmt.Errorf("UpdateDemands: commit: %w", err)
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
