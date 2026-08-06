package authenticate

import (
	"context"
	"fmt"
	"log"

	"dnd_schedule/internal/domain/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type IDatasource interface {
	AddUser(ctx context.Context, user *models.User) (*models.User, error)
	GetUser(ctx context.Context, username string) (*models.User, error)
}

type Datasource struct {
	pool *pgxpool.Pool
}

func NewAuthDatasource(pool *pgxpool.Pool) IDatasource {
	return &Datasource{pool: pool}
}

func (d Datasource) AddUser(ctx context.Context, user *models.User) (*models.User, error) {
	query := `INSERT INTO users (username, password_hash, created_at) 
	VALUES (@username, @password_hash, @created_at) RETURNING user_id`

	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not start transaction: %w", err)
	}

	var userID models.UserID
	if err := tx.QueryRow(ctx, query, user.ToNamedArgs()).Scan(&userID); err != nil {
		_ = tx.Rollback(ctx)
		return nil, fmt.Errorf("could not add user: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		_ = tx.Rollback(ctx)
		return nil, fmt.Errorf("could not commit transaction: %w", err)
	}
	log.Println("Successfully add user")
	user.ID = userID

	return user, nil
}

func (d Datasource) GetUser(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	query := `SELECT user_id, username, password_hash, created_at FROM users WHERE username = $1`
	err := d.pool.QueryRow(ctx, query, username).Scan(&user.ID, &user.Username, &user.Hash, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}
	log.Println("successfully retrieved user", user.Username)
	return &user, nil
}
