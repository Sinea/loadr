package loadr

import "errors"

type inMemoryStore struct {
	data map[Token]*Progress
}

func (s *inMemoryStore) Get(token Token) (*Progress, error) {
	if p, ok := s.data[token]; ok {
		return p, nil
	}

	return nil, errors.New("progress not found")
}

func (s *inMemoryStore) Set(token Token, progress *Progress) error {
	s.data[token] = progress
	return nil
}

func (s *inMemoryStore) Delete(token Token) error {
	delete(s.data, token)
	return nil
}

func newInMemoryStore() (ProgressStore, error) {
	return &inMemoryStore{
		data: make(map[Token]*Progress),
	}, nil
}
