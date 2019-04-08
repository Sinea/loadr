package loadr

import "time"

const (
	_ uint = iota
	Storage
	Broadcast

	// Channel error codes
	ChannelCloseError = 1 + iota
	ChannelSubscribeError
	ChannelUnmarshalError
)

type Error struct {
	Code    uint
	Message string
}

type Token string

// Subscription represents a client subscription on a token
type Subscription struct {
	Token  Token
	Client Client
}

// UpdateProgressRequest ...
type UpdateProgressRequest struct {
	Guarantee uint     `json:"guarantee" validate:"min=0"`
	Progress  Progress `json:"progress"`
}

// Progress ...
type Progress struct {
	Stage    string  `json:"stage" bson:"stage" validate:"min=1,max=200,regexp=^[a-zA-Z0-9]*$"`
	Progress float32 `json:"progress" bson:"progress" validate:"min=0,max=1"`
}

// MetaProgress bundle the progress with it's token to be sent and received over a Channel
type MetaProgress struct {
	Token    Token
	Progress Progress
}

// NetConfig used for backend and client listeners
type NetConfig struct {
	Address  string
	CertFile string
	KeyFile  string
}

// ErrorProvider ...
type ErrorProvider interface {
	Errors() <-chan error
}

// ProgressHandler ...
type ProgressHandler interface {
	Delete(token Token) error
	Set(token Token, progress *Progress, guarantee uint) error
}

// Service that dispatches progress
type Service interface {
	ErrorProvider
	ProgressHandler

	Run(backend BackendListener, clients ClientListener)
	SetCleanupInterval(duration time.Duration)
}

// ProgressStore interface for progress persistence
type ProgressStore interface {
	Get(token Token) (*Progress, error)
	Set(token Token, progress *Progress) error
	Delete(token Token) error
}

// Channel used to send progresses to other nodes
type Channel interface {
	ErrorProvider

	Push(MetaProgress) error
	Progresses() <-chan MetaProgress
	Close() error
}

// Client is a client that can receive progress updates
type Client interface {
	Write(progress *Progress) error
	Close() error
	IsAlive() bool
}

// ClientListener provides new clients that are interested in progress updates
type ClientListener interface {
	Wait() <-chan *Subscription
	Close() error
}

// BackendListener provides an interface for inputting progresses from backend
type BackendListener interface {
	Run(handler ProgressHandler)
}
