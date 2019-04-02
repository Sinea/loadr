package loadr

import (
	"encoding/json"
	"fmt"

	"github.com/garyburd/redigo/redis"
)

const queueName = "loadr"

type RedisConfig struct {
	Address string
}

type redisChannel struct {
	pool      *redis.Pool
	out       chan MetaProgress
	errors    chan error
	isRunning bool
}

func (r *redisChannel) Close() error {
	r.isRunning = false
	return nil
}

func (r *redisChannel) Push(p MetaProgress) error {
	connection := r.pool.Get()
	bytes, err := json.Marshal(p)
	if err != nil {
		return err
	}
	if _, err := connection.Do("PUBLISH", queueName, bytes); err != nil {
		return err
	}
	return nil
}

func (r *redisChannel) Progresses() <-chan MetaProgress {
	return r.out
}

func (r *redisChannel) Errors() <-chan error {
	return r.errors
}

func (r *redisChannel) read() {
	connection := r.pool.Get()
	// Cleanup crew
	defer func() {
		if err := connection.Close(); err != nil {
			r.errors <- fmt.Errorf("error closing connection: %s", err)
		}
	}()

	psc := redis.PubSubConn{Conn: connection}
	if err := psc.PSubscribe(queueName); err != nil {
		r.errors <- fmt.Errorf("error subscribing: %s", err)
		return
	}

	for r.isRunning {
		if message, ok := psc.Receive().(redis.Message); ok {
			p := MetaProgress{}
			if err := json.Unmarshal(message.Data, &p); err != nil {
				r.errors <- fmt.Errorf("error unmarshalling progress: %s", err)
			} else {
				r.out <- p
			}
		}
	}
}

func newRedisChannel(config RedisConfig) Channel {
	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", config.Address)
		},
	}

	result := &redisChannel{
		isRunning: true,
		pool:      pool,
		out:       make(chan MetaProgress),
		errors:    make(chan error),
	}

	go result.read()

	return result
}
