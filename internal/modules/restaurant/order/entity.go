package order

import "time"

type OrderStatus string

const (
	StatusAwaitingConfirmation OrderStatus = "awaiting_confirmation"
	StatusPending              OrderStatus = "pending"
	StatusConfirmed            OrderStatus = "confirmed"
	StatusPreparing            OrderStatus = "preparing"
	StatusReady                OrderStatus = "ready"
	StatusDelivered            OrderStatus = "delivered"
	StatusCancelled            OrderStatus = "cancelled"
)

type OrderConfirmation struct {
	ID        string
	OrderID   string
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type Order struct {
	ID           string
	RestaurantID string
	TableID      string
	Status       OrderStatus
	Notes        string
	GuestEmail   string
	Items        []OrderItem
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type OrderItem struct {
	ID       string
	OrderID  string
	Name     string
	Quantity int
	Notes    string
}
