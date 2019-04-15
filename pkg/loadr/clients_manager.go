package loadr

import (
	"log"
	"sync/atomic"
)

type clientsManager struct {
	isRunningCleaning uint32
	logger            *log.Logger
	clients           map[Token][]Client
}

// Send progress to all subscribed clients
func (c *clientsManager) Send(token Token, progress *Progress) {
	if clients, ok := c.clients[token]; ok {
		for _, client := range clients {
			if err := client.Write(progress); err != nil {
				c.logger.Printf("error writing to client: %s\n", err)
				c.closeClient(client)
			}
		}
	}
}

// ClearToken close all subscribed clients to this token
func (c *clientsManager) ClearToken(token Token) {
	if clients, ok := c.clients[token]; ok {
		for _, client := range clients {
			c.closeClient(client)
		}
	}
	delete(c.clients, token)
}

// AddClient to the clients map
func (c *clientsManager) AddClient(token Token, client Client) {
	c.clients[token] = append(c.clients[token], client)
}

// Cleanup dead client connections
func (c *clientsManager) CleanupClients() {
	if atomic.CompareAndSwapUint32(&c.isRunningCleaning, 0, 1) {
		return
	}
	defer atomic.CompareAndSwapUint32(&c.isRunningCleaning, 1, 0)
	// Cleanup clients one token at a time
	for token, clients := range c.clients {
		temp := c.cleanupTokenClients(clients)
		// TODO : Lock?
		c.clients[token] = temp
		// TODO : Unlock
	}
}

// cleanupTokenClients remove dead clients from the slice
func (c *clientsManager) cleanupTokenClients(clients []Client) []Client {
	remaining := make([]Client, 0)
	for _, cl := range clients {
		if cl.IsAlive() {
			remaining = append(remaining, cl)
		} else {
			c.closeClient(cl)
		}
	}

	return remaining
}

// closeClient and log the error, if any
func (c *clientsManager) closeClient(client Client) {
	if err := client.Close(); err != nil {
		c.logger.Println("error closing client socket")
	}
}
