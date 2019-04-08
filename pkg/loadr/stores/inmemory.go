package stores

import (
	"errors"

	"github.com/Sinea/loadr/pkg/loadr"
)

type inMemory struct {
	data map[loadr.Token]*loadr.Progress
}

func (s *inMemory) Get(token loadr.Token) (*loadr.Progress, error) {
	if p, ok := s.data[token]; ok {
		return p, nil
	}

	return nil, errors.New("progress not found")
}

func (s *inMemory) Set(token loadr.Token, progress *loadr.Progress) error {
	s.data[token] = progress
	return nil
}

func (s *inMemory) Delete(token loadr.Token) error {
	delete(s.data, token)
	return nil
}

func newInMemoryStore() (loadr.ProgressStore, error) {
	return &inMemory{
		data: make(map[loadr.Token]*loadr.Progress),
	}, nil
}
