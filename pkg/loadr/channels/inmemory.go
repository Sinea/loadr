package channels

import "github.com/Sinea/loadr/pkg/loadr"

type inMemory struct {
	out    chan loadr.MetaProgress
	errors chan error
}

func (c *inMemory) Errors() <-chan error {
	return c.errors
}

func (c *inMemory) Close() error {
	return nil
}

func (c *inMemory) Push(progress loadr.MetaProgress) error {
	c.out <- progress
	return nil
}

func (c *inMemory) Progresses() <-chan loadr.MetaProgress {
	return c.out
}

func newInMemoryChannel() loadr.Channel {
	return &inMemory{
		errors: make(chan error),
		out:    make(chan loadr.MetaProgress),
	}
}
