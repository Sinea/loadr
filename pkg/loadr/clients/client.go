package clients

import (
	"time"

	"github.com/Sinea/loadr/pkg/loadr"
	"github.com/gorilla/websocket"
)

type client struct {
	socket *websocket.Conn
}

func (c *client) IsAlive() bool {
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

func (c *client) Write(progress *loadr.Progress) error {
	return c.socket.WriteJSON(progress)
}

func (c *client) Close() error {
	return c.socket.Close()
}
