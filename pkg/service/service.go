package service

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"log"
)

var (
	upgrader = websocket.Upgrader{}
)

const (
	_ uint = iota
	Storage
	Broadcast
)

type Token string

type UpdateProgressRequest struct {
	Guarantee uint     `json:"guarantee"`
	Progress  Progress `json:"progress"`
}

type Progress struct {
	Stage    string  `json:"stage" bson:"stage"`
	Progress float32 `json:"progress" bson:"progress"`
}

type MetaProgress struct {
	Token    Token
	Progress Progress
}

type NetConfig struct {
	Address string
}

type Service interface {
	Listen(http, ws NetConfig) error
}

type service struct {
	store   ProgressStore
	channel Channel
	clients map[Token][]*websocket.Conn
}

func (s *service) Listen(backend, clients NetConfig) error {
	backendEndpoint := echo.New()
	backendEndpoint.POST("/:token", s.updateProgress)
	backendEndpoint.DELETE("/:token", s.deleteProgress)
	go func() {
		log.Printf("Listening for backend information on %s\n", backend.Address)
		backendEndpoint.Logger.Fatal(backendEndpoint.Start(backend.Address))
	}()

	e := echo.New()
	e.GET("/:token", s.wsHandler)
	go func() {
		log.Printf("Waiting for clients on %s\n", clients.Address)
		e.Logger.Fatal(e.Start(clients.Address))
	}()

	go func() {
		for {
			p := <-s.channel.Progresses()
			if clients, ok := s.clients[p.Token]; ok {
				for _, client := range clients {
					if err := client.WriteJSON(p.Progress); err != nil {
						log.Printf("error writing to client: %s\n", err)
						closeWebSocket(client)
					}
				}
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
			log.Printf("error writing initial progress state: %s\n", err)
			return err
		}
	} else {
		log.Printf("error retrieving initial progress state: %s\n", err)
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
			log.Printf("error saving progress: %s\n", err)
			return err
		}
	}

	if err := s.channel.Push(MetaProgress{token, update.Progress}); err != nil {
		if update.Guarantee >= Broadcast {
			log.Printf("error broadcasting progress: %s\n", err)
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
		return fmt.Errorf("error deleting progress for token '%s' : %s", token, err)
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
	}
}
