package models

type Slot struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	ValidFrom  string `json:"valid_from"`
	ValidUntil string `json:"valid_until"`
}
