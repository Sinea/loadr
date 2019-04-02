package loadr

func NewChannel(config interface{}) Channel {
	switch c := config.(type) {
	case RedisConfig:
		return newRedisChannel(c)
	default:
		return newInMemoryChannel()
	}
}
