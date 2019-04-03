package loadr

import (
	"io"
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
	clients         map[Token][]*websocket.Conn
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

	clientsEndpoint := echo.New()
	clientsEndpoint.GET("/:token", s.wsHandler)
	go startServer(clientsEndpoint, clients)

	go func() {
		ticker := time.NewTicker(s.cleanupInterval)
		for {
			select {
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
			if err := client.WriteJSON(p.Progress); err != nil {
				s.logger.Printf("error writing to client %s: %s\n", client.RemoteAddr(), err)
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

func (s *service) cleanupTokenClients(clients []*websocket.Conn) []*websocket.Conn {
	remaining := make([]*websocket.Conn, 0)
	for _, c := range clients {
		if err := c.SetReadDeadline(time.Now().Add(time.Millisecond)); err != nil {
			s.logger.Printf("error setting deadline %s", err)
			s.closeWebSocket(c)
			continue
		}
		if _, _, err := c.ReadMessage(); err != nil {
			if _, ok := err.(*websocket.CloseError); ok {
				s.closeWebSocket(c)
			} else {
				remaining = append(remaining, c)
			}
		} else {
			remaining = append(remaining, c)
		}
	}

	return remaining
}

func (s *service) wsHandler(c echo.Context) error {
	token := Token(c.Param("token"))
	ws, err := s.upgrader.Upgrade(c.Response(), c.Request(), nil)

	if err != nil {
		s.logger.Printf("error upgrading protocol: %s\n", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	if progress, err := s.store.Get(token); err == nil {
		if err := ws.WriteJSON(progress); err != nil {
			s.logger.Printf("error writing initial progress state: %s\n", err)
			s.closeWebSocket(ws)
			return c.NoContent(http.StatusInternalServerError)
		}
	} else {
		s.logger.Printf("error retrieving initial progress state: %s\n", err)
	}

	s.clients[token] = append(s.clients[token], ws)

	return nil
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

func (s *service) closeWebSocket(ws io.Closer) {
	if err := ws.Close(); err != nil {
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
