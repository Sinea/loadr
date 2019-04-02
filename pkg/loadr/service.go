package loadr

import (
	"fmt"
	"io"
	"log"
	"net/http"
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
}

func (s *service) Errors() <-chan error {
	return s.errors
}

func (s *service) Listen(backend, clients NetConfig) error {
	backendEndpoint := echo.New()
	backendEndpoint.POST("/:token", s.updateProgress)
	backendEndpoint.DELETE("/:token", s.deleteProgress)
	go func() {
		backendEndpoint.Logger.Fatal(backendEndpoint.Start(backend.Address))
	}()

	clientsEndpoint := echo.New()
	clientsEndpoint.GET("/:token", s.wsHandler)
	go func() {
		clientsEndpoint.Logger.Fatal(clientsEndpoint.Start(clients.Address))
	}()

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

// Handle an incoming progress
func (s *service) handleProgress(p MetaProgress) {
	if clients, ok := s.clients[p.Token]; ok {
		for _, client := range clients {
			if err := client.WriteJSON(p.Progress); err != nil {
				s.errors <- fmt.Errorf("error writing to client: %s", err)
				closeWebSocket(client)
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
	// TODO : Lock each token
	for token, clients := range s.clients {
		remaining := make([]*websocket.Conn, 0)
		for _, c := range clients {
			if err := c.SetReadDeadline(time.Now().Add(time.Millisecond)); err != nil {
				log.Printf("error setting deadline %s", err)
			}
			if _, _, err := c.ReadMessage(); err != nil && websocket.IsCloseError(err, websocket.CloseGoingAway) {
				closeWebSocket(c)
			} else {
				remaining = append(remaining, c)
			}
		}
		s.clients[token] = remaining
	}
	s.isCleaningUp = false
}

func (s *service) wsHandler(c echo.Context) error {
	token := Token(c.Param("token"))
	ws, err := s.upgrader.Upgrade(c.Response(), c.Request(), nil)

	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if progress, err := s.store.Get(token); err == nil {
		if err := ws.WriteJSON(progress); err != nil {
			s.errors <- fmt.Errorf("error writing initial progress state: %s", err)
			return c.String(http.StatusInternalServerError, err.Error())
		}
	} else {
		s.errors <- fmt.Errorf("error retrieving initial progress state: %s", err)
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
		return err
	}
	if err := s.store.Set(token, &update.Progress); err != nil {
		err := fmt.Errorf("error saving progress: %s", err)
		s.errors <- err
		if update.Guarantee >= Storage {
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}
	if err := s.channel.Push(MetaProgress{token, update.Progress}); err != nil {
		err := fmt.Errorf("error broadcasting progress: %s", err)
		s.errors <- err
		if update.Guarantee >= Broadcast {
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}
	return c.NoContent(http.StatusOK)
}

func (s *service) deleteProgress(c echo.Context) error {
	tokenString := c.Param("token")
	token := Token(tokenString)
	if clients, ok := s.clients[token]; ok {
		for _, client := range clients {
			closeWebSocket(client)
		}
	}
	s.clients[token] = make([]*websocket.Conn, 0)
	if err := s.store.Delete(token); err != nil {
		s.errors <- fmt.Errorf("error deleting progress for token '%s' : %s", token, err)
		return c.NoContent(http.StatusNotFound)
	}
	return c.NoContent(http.StatusOK)
}

func closeWebSocket(ws io.Closer) {
	if err := ws.Close(); err != nil {
		log.Println("error closing client socket")
	}
}

func New(store ProgressStore, channel Channel) Service {
	return &service{
		cleanupInterval: time.Second * 3,
		upgrader:        websocket.Upgrader{},
		store:           store,
		channel:         channel,
		clients:         map[Token][]*websocket.Conn{},
		errors:          make(chan error),
	}
}
