package menu

import "time"

type Menu struct {
	ID           string
	RestaurantID string
	Name         string
	Description  string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type MenuItem struct {
	ID          string
	MenuID      string
	Name        string
	Description string
	Price       float64
	Position    int
	CreatedAt   time.Time
}
