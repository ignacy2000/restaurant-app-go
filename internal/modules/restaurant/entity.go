package restaurant

import "time"

type Restaurant struct {
	ID          string
	UserID      string
	Name        string
	Description string
	Address     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
