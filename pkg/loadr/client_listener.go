package loadr

import (
	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"log"
	"net/http"
)

type websocketListener struct {
	upgrader websocket.Upgrader
	logger   *log.Logger
	clients  chan *clientSubscription
}

func (c *websocketListener) Wait(config NetConfig) <-chan *clientSubscription {
	clientsEndpoint := echo.New()
	clientsEndpoint.GET("/:token", c.wsHandler)
	go startServer(clientsEndpoint, config)

	return c.clients
}

func (c *websocketListener) wsHandler(ctx echo.Context) error {
	token := Token(ctx.Param("token"))
	connection, err := c.upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)

	if err != nil {
		c.logger.Printf("error upgrading protocol: %s\n", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	c.clients <- &clientSubscription{
		Token: token,
		Client: &client{
			socket: connection,
		},
	}

	return nil
}

func newClientListener(logger *log.Logger) clientListener {
	return &websocketListener{logger: logger}
}
