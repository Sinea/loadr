package loadr

type Channel interface {
	Push(MetaProgress) error
	Progresses() <-chan MetaProgress
	Errors() <-chan error
	Close() error
}

func NewChannel(config interface{}) Channel {
	switch c := config.(type) {
	case RedisConfig:
		return newRedisChannel(c)
	default:
		return &inMemoryChannel{
			errors: make(chan error),
			out:    make(chan MetaProgress),
		}
	}
}
