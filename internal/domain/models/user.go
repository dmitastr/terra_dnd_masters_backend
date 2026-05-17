package models

import (
	"time"

	"github.com/jackc/pgx/v5"
)

type UserID int

type User struct {
	Username  string    `json:"username" db:"username"`
	ID        UserID    `json:"user_id" db:"user_id"`
	Password  string    `json:"password"`
	Hash      string    `json:"hash" db:"password_hash"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func (u *User) ToNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{"username": u.Username, "password_hash": u.Hash, "created_at": time.Now()}
}
