package loadr

type inMemoryChannel struct {
	out    chan MetaProgress
	errors chan error
}

func (c *inMemoryChannel) Errors() <-chan error {
	return c.errors
}

func (c *inMemoryChannel) Close() error {
	return nil
}

func (c *inMemoryChannel) Push(progress MetaProgress) error {
	c.out <- progress
	return nil
}

func (c *inMemoryChannel) Progresses() <-chan MetaProgress {
	return c.out
}
