package ws

import (
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = pongWait * 9 / 10
	maxMessageSize = 1024
)

// Client reprezentuje pojedyncze połączenie WebSocket.
// restaurantID ustawiony → właściciel restauracji.
// tableID ustawiony → klient przy stoliku.
type Client struct {
	hub          *Hub
	conn         *websocket.Conn
	send         chan []byte
	restaurantID string
	tableID      string
}

func newOwnerClient(hub *Hub, conn *websocket.Conn, restaurantID string) *Client {
	return &Client{hub: hub, conn: conn, send: make(chan []byte, 256), restaurantID: restaurantID}
}

func newTableClient(hub *Hub, conn *websocket.Conn, tableID string) *Client {
	return &Client{hub: hub, conn: conn, send: make(chan []byte, 256), tableID: tableID}
}

// start uruchamia write pump w tle i blokuje na read pump do rozłączenia.
func (c *Client) start() {
	hub := c.hub
	hub.register <- c
	go c.writePump()
	c.readPump()
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	// klient WS tylko odbiera — czytamy ale ignorujemy wiadomości przychodzące
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}
