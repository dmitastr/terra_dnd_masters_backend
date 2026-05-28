package models

import "time"

type Master struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	VkID      int       `json:"vk_id" db:"vk_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"created_at" db:"updated_at"`
}
