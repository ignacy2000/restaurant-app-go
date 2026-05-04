package ws

type EventType string

const (
	EventOrderCreated       EventType = "order.created"
	EventOrderStatusChanged EventType = "order.status_changed"
	EventCallCreated        EventType = "call.created"
	EventCallStatusChanged  EventType = "call.status_changed"
)

type Event struct {
	Type    EventType `json:"type"`
	Payload any       `json:"payload"`
}
