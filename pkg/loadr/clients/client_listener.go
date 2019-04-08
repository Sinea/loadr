package clients

import (
	"log"
	"net/http"

	"github.com/Sinea/loadr/pkg/loadr"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
)

type clientListener struct {
	config   loadr.NetConfig
	endpoint *echo.Echo
	upgrader websocket.Upgrader
	logger   *log.Logger
	clients  chan *loadr.Subscription
}

func (c *clientListener) Wait() <-chan *loadr.Subscription {
	if c.endpoint == nil {
		c.endpoint = echo.New()
		c.endpoint.GET("/:token", c.websocketHandler)
		go startServer(c.endpoint, c.config)
	}
	return c.clients
}

func (c *clientListener) Close() error {
	err := c.endpoint.Close()
	c.endpoint = nil
	return err
}

func (c *clientListener) websocketHandler(ctx echo.Context) error {
	token := loadr.Token(ctx.Param("token"))
	connection, err := c.upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)

	if err != nil {
		c.logger.Printf("error upgrading protocol: %s\n", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	c.clients <- &loadr.Subscription{
		Token: token,
		Client: &client{
			socket: connection,
		},
	}

	return nil
}

func New(config loadr.NetConfig, logger *log.Logger) loadr.ClientListener {
	return &clientListener{
		clients:  make(chan *loadr.Subscription),
		config:   config,
		logger:   logger,
		upgrader: websocket.Upgrader{},
	}
}

func startServer(server *echo.Echo, config loadr.NetConfig) {
	var err error
	if config.KeyFile != "" && config.CertFile != "" {
		err = server.StartTLS(config.Address, config.CertFile, config.KeyFile)
	} else {
		err = server.Start(config.Address)
	}

	server.Logger.Fatal(err)
}
