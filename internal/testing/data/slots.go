package data

import (
	"dnd_schedule/internal/domain/models"
)

var Slots = []models.Slot{
	{ID: 13, Name: "пн вечер", ValidFrom: "2020-01-01", ValidUntil: "9999-12-31"},
	{ID: 23, Name: "вт вечер", ValidFrom: "2020-01-01", ValidUntil: "9999-12-31"},
	{ID: 33, Name: "ср вечер", ValidFrom: "2020-01-01", ValidUntil: "9999-12-31"},
	{ID: 61, Name: "сб утро", ValidFrom: "2020-01-01", ValidUntil: "9999-12-31"},
	{ID: 63, Name: "сб вечер", ValidFrom: "2020-01-01", ValidUntil: "9999-12-31"},
}
