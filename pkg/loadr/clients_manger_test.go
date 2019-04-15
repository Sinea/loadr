package loadr

import (
	"errors"
	"log"
	"os"
	"testing"
)

func TestClientsManager_CleanupClients(t *testing.T) {
	managher := &clientsManager{
		logger:  log.New(os.Stdout, "", 0),
		clients: make(map[Token][]Client),
	}

	clientA := &mockClient{}
	clientA.On("Close").Once().Return(errors.New("close error"))
	clientA.On("IsAlive").Once().Return(false)
	clientB := &mockClient{}
	clientB.On("Close").Once().Return(errors.New("close error"))
	clientB.On("IsAlive").Once().Return(true)

	managher.AddClient(Token("x"), clientA)
	managher.AddClient(Token("x"), clientB)

	managher.CleanupClients()
	managher.CleanupClients()
}
