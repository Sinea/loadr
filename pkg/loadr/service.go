package loadr

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"gopkg.in/validator.v2"
)

type service struct {
	upgrader        websocket.Upgrader
	store           ProgressStore
	channel         Channel
	clients         map[Token][]Client
	errors          chan error
	cleanupInterval time.Duration
	isCleaningUp    bool
	logger          *log.Logger
}

func (s *service) SetCleanupInterval(duration time.Duration) {
	s.cleanupInterval = duration
}

func (s *service) Errors() <-chan error {
	return s.errors
}

func (s *service) Listen(backend, clients NetConfig) error {
	backendEndpoint := echo.New()
	backendEndpoint.POST("/:token", s.updateProgress)
	backendEndpoint.DELETE("/:token", s.deleteProgress)
	go startServer(backendEndpoint, backend)

	listener := newClientListener(s.logger)

	go func() {
		ticker := time.NewTicker(s.cleanupInterval)
		for {
			select {
			case c := <-listener.Wait(clients):
				s.wsHandler(c)
			case p := <-s.channel.Progresses():
				s.handleProgress(p)
			case err := <-s.channel.Errors():
				s.errors <- err
			case <-ticker.C:
				go s.cleanupClients()
			}
		}
	}()

	return nil
}

func startServer(server *echo.Echo, config NetConfig) {
	var err error
	if config.KeyFile != "" && config.CertFile != "" {
		err = server.StartTLS(config.Address, config.CertFile, config.KeyFile)
	} else {
		err = server.Start(config.Address)
	}

	server.Logger.Fatal(err)
}

// Handle an incoming progress
func (s *service) handleProgress(p MetaProgress) {
	if clients, ok := s.clients[p.Token]; ok {
		for _, client := range clients {
			if err := client.Write(&p.Progress); err != nil {
				s.logger.Printf("error writing to client: %s\n", err)
				s.closeWebSocket(client)
			}
		}
	}
}

// Cleanup client connections
func (s *service) cleanupClients() {
	if s.isCleaningUp {
		return
	}
	s.isCleaningUp = true
	for token, clients := range s.clients {
		// TODO : Lock
		s.clients[token] = s.cleanupTokenClients(clients)
	}
	s.isCleaningUp = false
}

func (s *service) cleanupTokenClients(clients []Client) []Client {
	remaining := make([]Client, 0)
	for _, c := range clients {
		if !c.IsAlive() {
			remaining = append(remaining, c)
		} else {
			s.closeWebSocket(c)
		}
	}

	return remaining
}

func (s *service) wsHandler(client Client) {
	token := client.Token()

	if progress, err := s.store.Get(token); err == nil {
		if err := client.Write(progress); err != nil {
			s.logger.Printf("error writing initial progress state: %s\n", err)
			s.closeWebSocket(client)
		}
	} else {
		s.logger.Printf("error retrieving initial progress state: %s\n", err)
	}

	s.clients[token] = append(s.clients[token], client)
}

func (s *service) updateProgress(c echo.Context) error {
	tokenString := c.Param("token")
	token := Token(tokenString)
	update := &UpdateProgressRequest{}

	if err := c.Bind(update); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if err := validator.Validate(update); err != nil {
		s.logger.Printf("error validating progress request: %s\n", err)
		return c.NoContent(http.StatusBadRequest)
	}
	if err := s.store.Set(token, &update.Progress); err != nil {
		s.logger.Printf("error saving progress: %s\n", err)
		if update.Guarantee >= Storage {
			return c.NoContent(http.StatusInternalServerError)
		}
	}
	if err := s.channel.Push(MetaProgress{token, update.Progress}); err != nil {
		s.logger.Printf("error broadcasting progress: %s\n", err)
		if update.Guarantee >= Broadcast {
			return c.NoContent(http.StatusInternalServerError)
		}
	}
	return c.NoContent(http.StatusOK)
}

func (s *service) deleteProgress(c echo.Context) error {
	tokenString := c.Param("token")
	token := Token(tokenString)
	if clients, ok := s.clients[token]; ok {
		for _, client := range clients {
			s.closeWebSocket(client)
		}
	}
	s.clients[token] = make([]*websocket.Conn, 0)
	if err := s.store.Delete(token); err != nil {
		s.logger.Printf("error deleting progress for token '%s' : %s\n", token, err)
		return c.NoContent(http.StatusNotFound)
	}
	return c.NoContent(http.StatusOK)
}

func (s *service) closeWebSocket(client Client) {
	if err := client.Close(); err != nil {
		s.logger.Println("error closing client socket")
	}
}

func New(store ProgressStore, channel Channel) Service {
	return &service{
		logger:          log.New(os.Stdout, "", 0),
		cleanupInterval: time.Second * 30,
		upgrader:        websocket.Upgrader{},
		store:           store,
		channel:         channel,
		clients:         map[Token][]*websocket.Conn{},
		errors:          make(chan error),
	}
}
