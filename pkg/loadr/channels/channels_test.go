package channels

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNew_InMemoryCreation(t *testing.T) {
	_, ok := New(nil).(*inMemory)
	assert.True(t, ok)
}

func TestNew_RedisChannel(t *testing.T) {
	_, ok := New(RedisConfig{}).(*redisChannel)
	assert.True(t, ok)
}
