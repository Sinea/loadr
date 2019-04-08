package channels

import (
	"encoding/json"
	"fmt"

	"github.com/Sinea/loadr/pkg/loadr"
	"github.com/garyburd/redigo/redis"
)

type RedisConfig struct {
	Address string
}

type redisChannel struct {
	pool      *redis.Pool
	out       chan loadr.MetaProgress
	errors    chan error
	isRunning bool
	queueName string
}

func (r *redisChannel) Close() error {
	r.isRunning = false
	return nil
}

func (r *redisChannel) Push(p loadr.MetaProgress) error {
	connection := r.pool.Get()
	bytes, err := json.Marshal(p)
	if err != nil {
		return err
	}
	if _, err := connection.Do("PUBLISH", r.queueName, bytes); err != nil {
		return err
	}
	return nil
}

func (r *redisChannel) Progresses() <-chan loadr.MetaProgress {
	return r.out
}

func (r *redisChannel) Errors() <-chan error {
	return r.errors
}

func (r *redisChannel) read() {
	connection := r.pool.Get()
	defer r.closeConnection(connection)

	subscription, err := r.subscribe(connection, r.queueName)

	if err != nil {
		r.errors <- &loadr.Error{
			Message: fmt.Sprintf("error subscribing: %s", err),
			Code:    loadr.ChannelSubscribeError,
		}
		return
	}

	r.receive(subscription)
}

func (r *redisChannel) receive(subscription *redis.PubSubConn) {
	defer r.closeSubscription(subscription)

	for r.isRunning {
		r.readMessage(subscription)
	}
}

func (r *redisChannel) readMessage(subscription *redis.PubSubConn) {
	if message, ok := subscription.Receive().(redis.Message); ok {
		p := loadr.MetaProgress{}
		if err := json.Unmarshal(message.Data, &p); err != nil {
			r.errors <- &loadr.Error{
				Message: fmt.Sprintf("error unmarshalling progress: %s", err),
				Code:    loadr.ChannelUnmarshalError,
			}
		} else {
			r.out <- p
		}
	}
}

func (r *redisChannel) closeSubscription(subscription *redis.PubSubConn) {
	if err := subscription.Close(); err != nil {
		r.errors <- &loadr.Error{
			Message: fmt.Sprintf("error unsubscribing: %s", err),
			Code:    loadr.ChannelCloseError,
		}
	}
}

func (r *redisChannel) closeConnection(connection redis.Conn) {
	if err := connection.Close(); err != nil {
		r.errors <- &loadr.Error{
			Message: fmt.Sprintf("error closing connection: %s", err),
			Code:    loadr.ChannelCloseError,
		}
	}
}

func (r *redisChannel) subscribe(connection redis.Conn, queue string) (*redis.PubSubConn, error) {
	subscription := &redis.PubSubConn{Conn: connection}
	if err := subscription.PSubscribe(queue); err != nil {
		return nil, err
	}
	return subscription, nil
}

func newRedisPool(address string) *redis.Pool {
	return &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", address)
		},
	}
}

func newRedisChannel(config RedisConfig) loadr.Channel {
	pool := newRedisPool(config.Address)

	result := &redisChannel{
		queueName: "loadr",
		isRunning: true,
		pool:      pool,
		out:       make(chan loadr.MetaProgress),
		errors:    make(chan error),
	}

	go result.read()

	return result
}
