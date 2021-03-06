package loadr

import (
	"fmt"
	"log"
	"time"

	"gopkg.in/validator.v2"
)

type service struct {
	store           Store
	channel         Channel
	clients         map[Token][]Client
	errors          chan error
	cleanupInterval time.Duration
	isCleaningUp    bool
	logger          *log.Logger
}

// Delete delete the progress for a specific token
func (s *service) Delete(token Token) error {
	if clients, ok := s.clients[token]; ok {
		for _, client := range clients {
			s.closeClient(client)
		}
	}
	s.clients[token] = make([]Client, 0)
	if err := s.store.Delete(token); err != nil {
		err := fmt.Errorf("error deleting progress for token '%s' : %s", token, err)
		s.logger.Println(err)
		return err
	}

	return nil
}

// Set update the progress for a token
func (s *service) Set(token Token, progress *Progress, guarantee uint) error {
	if err := validator.Validate(progress); err != nil {
		err := fmt.Errorf("error validating progress: %s", err)
		s.logger.Println(err)
		return err
	}
	if err := s.store.Set(token, progress); err != nil {
		err := fmt.Errorf("error saving progress: %s", err)
		s.logger.Println(err)
		if guarantee >= Storage {
			return err
		}
	}
	if err := s.channel.Push(MetaProgress{token, *progress}); err != nil {
		err := fmt.Errorf("error broadcasting progress: %s", err)
		s.logger.Println(err)
		if guarantee >= Broadcast {
			return err
		}
	}
	return nil
}

// Run the service
func (s *service) Run(backend BackendListener, clients ClientListener) {
	// Listen for backend progress information
	go backend.Run(s)

	go func() {
		ticker := time.NewTicker(s.cleanupInterval)
		for {
			select {
			case subscription := <-clients.Wait():
				s.HandleSubscription(subscription)
			case p := <-s.channel.Progresses():
				s.HandleProgress(p)
			case err := <-s.channel.Errors():
				s.errors <- err
			case <-ticker.C:
				go s.cleanupClients()
			}
		}
	}()
}

// SetCleanupInterval interval at which to clean up broken clients
func (s *service) SetCleanupInterval(duration time.Duration) {
	s.cleanupInterval = duration
}

// Errors produced by the service
func (s *service) Errors() <-chan error {
	return s.errors
}

// Handle an incoming progress
func (s *service) HandleProgress(progress MetaProgress) {
	if clients, ok := s.clients[progress.Token]; ok {
		for _, client := range clients {
			if err := client.Write(&progress.Progress); err != nil {
				s.logger.Printf("error writing to client: %s\n", err)
				s.closeClient(client)
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
			s.closeClient(c)
		}
	}

	return remaining
}

func (s *service) HandleSubscription(subscription *Subscription) {
	token := subscription.Token

	if progress, err := s.store.Get(token); err == nil {
		if err := subscription.Client.Write(progress); err != nil {
			s.logger.Printf("error writing initial progress state: %s\n", err)
			s.closeClient(subscription.Client)
		}
	} else {
		s.logger.Printf("error retrieving initial progress state: %s\n", err)
	}

	s.clients[token] = append(s.clients[token], subscription.Client)
}

// closeClient and log the error, if any
func (s *service) closeClient(client Client) {
	if err := client.Close(); err != nil {
		s.logger.Println("error closing client socket")
	}
}

// New service
func New(store Store, channel Channel, logger *log.Logger) Service {
	return &service{
		logger:          logger,
		cleanupInterval: time.Second * 30,
		store:           store,
		channel:         channel,
		clients:         make(map[Token][]Client),
		errors:          make(chan error),
	}
}
