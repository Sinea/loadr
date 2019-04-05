package loadr

import (
	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"log"
	"net/http"
)

type clientListener struct {
	upgrader websocket.Upgrader
	logger   *log.Logger
	clients  chan Client
}

func (c *clientListener) Wait(config NetConfig) <-chan Client {
	clientsEndpoint := echo.New()
	clientsEndpoint.GET("/:token", c.wsHandler)
	go startServer(clientsEndpoint, config)

	return c.clients
}

func (c *clientListener) wsHandler(ctx echo.Context) error {
	token := Token(ctx.Param("token"))
	connection, err := c.upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)

	if err != nil {
		c.logger.Printf("error upgrading protocol: %s\n", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	c.clients <- &client{socket: connection, token: token}

	return nil
}

func newClientListener(logger *log.Logger) ClientListener {
	return &clientListener{logger: logger}
}
