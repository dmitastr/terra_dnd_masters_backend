package models

import "time"

type Slot struct {
	ID         int       `json:"id" db:"id" example:"13"`
	Name       string    `json:"name" db:"name" example:"пн вечер"`
	ValidFrom  time.Time `json:"valid_from" db:"valid_from" example:"2020-01-01"`
	ValidUntil time.Time `json:"valid_until" db:"valid_until" example:"2020-01-01"`
}
