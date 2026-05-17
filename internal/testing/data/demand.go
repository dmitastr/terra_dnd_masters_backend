package data

import (
	"sync"
	"time"

	"dnd_schedule/internal/domain/models"
)

var demands = []models.Demand{
	{
		Slots: []models.Slot{
			{
				ID:         13,
				Name:       "пн вечер",
				ValidFrom:  "2020-01-01",
				ValidUntil: "9999-01-01",
			},
			{
				ID:         23,
				Name:       "вт вечер",
				ValidFrom:  "2020-01-01",
				ValidUntil: "9999-01-01",
			},
		},
		VkID:         0,
		FirstName:    "Иван",
		LastName:     "Пупкин",
		ForWeek:      week,
		PlayersCount: 1,
	},
	{
		Slots: []models.Slot{
			{
				ID:         61,
				Name:       "сб утро",
				ValidFrom:  "2020-01-01",
				ValidUntil: "9999-01-01",
			},
		},
		VkID:         0,
		FirstName:    "Мария",
		LastName:     "Залупкина",
		ForWeek:      week,
		PlayersCount: 2,
	},
}

var week, _ = time.Parse("2006-01-02", "2026-05-04")

type DemandsDS struct {
	mu      sync.Mutex
	demands []models.Demand
}

func NewDemandsDS() *DemandsDS {
	return &DemandsDS{demands: demands}
}

func (db *DemandsDS) GetDemands() []models.Demand {
	return db.demands
}

func (db *DemandsDS) AddDemands(demands []models.Demand) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.demands = append(db.demands, demands...)
}
