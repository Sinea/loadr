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
	clients         clientsBucket
	errors          chan error
	cleanupInterval time.Duration
	logger          *log.Logger
}

// Delete delete the progress for a specific token
func (s *service) Delete(token Token) error {
	s.clients.ClearToken(token)

	if err := s.store.Delete(token); err != nil {
		return fmt.Errorf("error deleting progress for token '%s' : %s", token, err)
	}

	return nil
}

// Set update the progress for a token
func (s *service) Set(token Token, progress *Progress, guarantee uint) error {
	if err := validator.Validate(progress); err != nil {
		return fmt.Errorf("error validating progress: %s", err)
	}
	if err := s.store.Set(token, progress); err != nil {
		if guarantee >= Storage {
			return fmt.Errorf("error saving progress: %s", err)
		}
	}
	if err := s.channel.Push(MetaProgress{token, *progress}); err != nil {
		if guarantee >= Broadcast {
			return fmt.Errorf("error broadcasting progress: %s", err)
		}
	}
	return nil
}

// Handle an incoming progress
func (s *service) HandleProgress(progress MetaProgress) {
	s.clients.Send(progress.Token, &progress.Progress)
}

// Subscribe a client to given token
func (s *service) Subscribe(token Token, client Client) {
	if progress, err := s.store.Get(token); err == nil {
		if err := client.Write(progress); err != nil {
			s.logger.Printf("error writing initial progress state: %s\n", err)
		}
	} else {
		s.logger.Printf("error retrieving initial progress state: %s\n", err)
	}

	s.clients.AddClient(token, client)
}

// New service
func New(store Store, channel Channel, logger *log.Logger) Service {
	return &service{
		logger:  logger,
		store:   store,
		channel: channel,
		clients: clientsBucket{
			logger:  logger,
			clients: make(map[Token][]Client),
		},
		errors: make(chan error),
	}
}
