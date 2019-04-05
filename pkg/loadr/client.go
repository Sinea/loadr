package loadr

import (
	"github.com/gorilla/websocket"
	"time"
)

type wsClient struct {
	socket *websocket.Conn
}

func (c *wsClient) IsAlive() bool {
	if err := c.socket.SetReadDeadline(time.Now().Add(time.Millisecond)); err != nil {
		return false
	}
	if _, _, err := c.socket.ReadMessage(); err != nil {
		if _, ok := err.(*websocket.CloseError); ok {
			return false
		}
	}

	return true
}

func (c *wsClient) Write(progress *Progress) error {
	return c.socket.WriteJSON(progress)
}

func (c *wsClient) Close() error {
	return c.socket.Close()
}
