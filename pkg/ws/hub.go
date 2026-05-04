package ws

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

type broadcastMsg struct {
	restaurantID string
	tableID      string
	data         []byte
}

// Hub zarządza połączeniami WebSocket.
// Wszystkie operacje na mapach wykonuje pojedyncza goroutine Run().
type Hub struct {
	register   chan *Client
	unregister chan *Client
	broadcast  chan broadcastMsg

	restaurantConns map[string]map[*Client]struct{}
	tableConns      map[string]map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{
		register:        make(chan *Client, 64),
		unregister:      make(chan *Client, 64),
		broadcast:       make(chan broadcastMsg, 512),
		restaurantConns: make(map[string]map[*Client]struct{}),
		tableConns:      make(map[string]map[*Client]struct{}),
	}
}

// Run musi być uruchomiony jako goroutine: go hub.Run()
func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			if c.restaurantID != "" {
				if h.restaurantConns[c.restaurantID] == nil {
					h.restaurantConns[c.restaurantID] = make(map[*Client]struct{})
				}
				h.restaurantConns[c.restaurantID][c] = struct{}{}
			}
			if c.tableID != "" {
				if h.tableConns[c.tableID] == nil {
					h.tableConns[c.tableID] = make(map[*Client]struct{})
				}
				h.tableConns[c.tableID][c] = struct{}{}
			}

		case c := <-h.unregister:
			if c.restaurantID != "" {
				if conns, ok := h.restaurantConns[c.restaurantID]; ok {
					delete(conns, c)
					if len(conns) == 0 {
						delete(h.restaurantConns, c.restaurantID)
					}
				}
			}
			if c.tableID != "" {
				if conns, ok := h.tableConns[c.tableID]; ok {
					delete(conns, c)
					if len(conns) == 0 {
						delete(h.tableConns, c.tableID)
					}
				}
			}
			close(c.send)

		case msg := <-h.broadcast:
			sent := make(map[*Client]struct{})
			if msg.restaurantID != "" {
				for c := range h.restaurantConns[msg.restaurantID] {
					h.sendTo(c, msg.data)
					sent[c] = struct{}{}
				}
			}
			if msg.tableID != "" {
				for c := range h.tableConns[msg.tableID] {
					if _, already := sent[c]; !already {
						h.sendTo(c, msg.data)
					}
				}
			}
		}
	}
}

func (h *Hub) sendTo(c *Client, data []byte) {
	select {
	case c.send <- data:
	default:
	}
}

func (h *Hub) BroadcastToRestaurant(restaurantID string, event Event) {
	data, _ := json.Marshal(event)
	select {
	case h.broadcast <- broadcastMsg{restaurantID: restaurantID, data: data}:
	default:
	}
}

func (h *Hub) BroadcastToTable(tableID string, event Event) {
	data, _ := json.Marshal(event)
	select {
	case h.broadcast <- broadcastMsg{tableID: tableID, data: data}:
	default:
	}
}

func (h *Hub) BroadcastToRestaurantAndTable(restaurantID, tableID string, event Event) {
	data, _ := json.Marshal(event)
	select {
	case h.broadcast <- broadcastMsg{restaurantID: restaurantID, tableID: tableID, data: data}:
	default:
	}
}

func (h *Hub) ServeOwner(conn *websocket.Conn, restaurantID string) {
	newOwnerClient(h, conn, restaurantID).start()
}

func (h *Hub) ServeTable(conn *websocket.Conn, tableID string) {
	newTableClient(h, conn, tableID).start()
}
