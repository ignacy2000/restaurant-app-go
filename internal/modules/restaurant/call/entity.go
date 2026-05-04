package call

import "time"

type CallStatus string

const (
	StatusPending      CallStatus = "pending"
	StatusAcknowledged CallStatus = "acknowledged"
	StatusDone         CallStatus = "done"
)

type WaiterCall struct {
	ID           string
	RestaurantID string
	TableID      string
	Status       CallStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
