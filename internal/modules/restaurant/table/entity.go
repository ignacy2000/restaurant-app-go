package table

import "time"

type Table struct {
	ID           string
	RestaurantID string
	Number       int
	Capacity     int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
