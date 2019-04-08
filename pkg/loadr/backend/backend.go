package backend

import (
	"net/http"

	"github.com/Sinea/loadr/pkg/loadr"
	"github.com/labstack/echo"
)

type backend struct {
	config  loadr.NetConfig
	handler loadr.ProgressHandler
}

func (b *backend) Run(handler loadr.ProgressHandler) {
	b.handler = handler
	endpoint := echo.New()
	endpoint.POST("/:token", b.updateProgress)
	endpoint.DELETE("/:token", b.deleteProgress)
	go startServer(endpoint, b.config)
}

func (b *backend) updateProgress(c echo.Context) error {
	tokenString := c.Param("token")
	token := loadr.Token(tokenString)
	update := &loadr.UpdateProgressRequest{}

	if err := c.Bind(update); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	if err := b.handler.Set(token, &update.Progress, update.Guarantee); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

func (b *backend) deleteProgress(c echo.Context) error {
	tokenString := c.Param("token")
	token := loadr.Token(tokenString)

	if err := b.handler.Delete(token); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
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

func New(config loadr.NetConfig) loadr.BackendListener {
	return &backend{
		config: config,
	}
}
