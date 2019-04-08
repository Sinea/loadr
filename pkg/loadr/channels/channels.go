package channels

import "github.com/Sinea/loadr/pkg/loadr"

func New(config interface{}) loadr.Channel {
	switch c := config.(type) {
	case RedisConfig:
		return newRedisChannel(c)
	default:
		return newInMemoryChannel()
	}
}
