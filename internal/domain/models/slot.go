package models

import "time"

type Slot struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	ValidFrom  time.Time `json:"valid_from"`
	ValidUntil time.Time `json:"valid_until"`
}
