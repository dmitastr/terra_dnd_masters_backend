package models

import "time"

type Demand struct {
	Slots     []Slot    `json:"slots"`
	VkID      int       `json:"vk_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	ForWeek   time.Time `json:"for_week"`
}
