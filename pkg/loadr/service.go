package loadr

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"log"
	"net/http"
)

var (
	upgrader = websocket.Upgrader{}
)

type service struct {
	store   ProgressStore
	channel Channel
	clients map[Token][]*websocket.Conn
	errors  chan error
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

	e := echo.New()
	e.GET("/:token", s.wsHandler)
	go func() {
		e.Logger.Fatal(e.Start(clients.Address))
	}()

	go func() {
		for {
			select {
			case p := <-s.channel.Progresses():
				if clients, ok := s.clients[p.Token]; ok {
					for _, client := range clients {
						if err := client.WriteJSON(p.Progress); err != nil {
							s.errors <- fmt.Errorf("error writing to client: %s", err)
							closeWebSocket(client)
						}
					}
				}
			case err := <-s.channel.Errors():
				s.errors <- err
			}
		}
	}()

	return nil
}

func (s *service) wsHandler(c echo.Context) error {
	token := Token(c.Param("token"))
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	if progress, err := s.store.Get(token); err == nil {
		if err := ws.WriteJSON(progress); err != nil {
			s.errors <- fmt.Errorf("error writing initial progress state: %s", err)
			return err
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
		log.Println(err)
	}

	if err := s.store.Set(token, &update.Progress); err != nil {
		if update.Guarantee >= Storage {
			s.errors <- fmt.Errorf("error saving progress: %s", err)
			return err
		}
	}

	if err := s.channel.Push(MetaProgress{token, update.Progress}); err != nil {
		if update.Guarantee >= Broadcast {
			s.errors <- fmt.Errorf("error broadcasting progress: %s", err)
			return err
		}
	}

	return nil
}

func (s *service) deleteProgress(c echo.Context) error {
	token := c.Param("token")
	if clients, ok := s.clients[Token(token)]; ok {
		for _, client := range clients {
			closeWebSocket(client)
		}
	}
	if err := s.store.Delete(Token(token)); err != nil {
		s.errors <- fmt.Errorf("error deleting progress for token '%s' : %s", token, err)
		return c.NoContent(http.StatusNotFound)
	}
	return nil
}

func closeWebSocket(ws *websocket.Conn) {
	if err := ws.Close(); err != nil {
		log.Println("error closing the socket")
	}
}

func New(store ProgressStore, channel Channel) Service {
	return &service{
		store:   store,
		channel: channel,
		clients: map[Token][]*websocket.Conn{},
		errors:  make(chan error),
	}
}
