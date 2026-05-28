package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type Demand struct {
	ID           int       `json:"id" db:"id"`
	Slots        []Slot    `json:"slots" db:"slots"`
	VkID         int       `json:"vk_id" db:"vk_id"`
	FirstName    string    `json:"first_name" db:"first_name"`
	LastName     string    `json:"last_name" db:"last_name"`
	ForWeek      time.Time `json:"for_week" db:"for_week"`
	PlayersCount int       `json:"players_count" db:"players_count"`
}

func (d *Demand) IsValid() bool {
	return d.VkID > 0 && d.PlayersCount > 0
}

func (d *Demand) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("invalid type")
	}

	return json.Unmarshal(b, d)
}

func (d *Demand) Value() (driver.Value, error) {
	return json.Marshal(d)
}
