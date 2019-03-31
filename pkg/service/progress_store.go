package service

import "errors"

type ProgressStore interface {
	Get(token Token) (*Progress, error)
	Set(token Token, progress *Progress) error
	Delete(token Token) error
}

type inMemory struct {
	data map[Token]*Progress
}

func (s *inMemory) Get(token Token) (*Progress, error) {
	if p, ok := s.data[token]; ok {
		return p, nil
	}

	return nil, errors.New("progress not found")
}

func (s *inMemory) Set(token Token, progress *Progress) error {
	s.data[token] = progress
	return nil
}

func (s *inMemory) Delete(token Token) error {
	delete(s.data, token)
	return nil
}
