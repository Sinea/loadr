package loadr

const (
	_ uint = iota
	Storage
	Broadcast
)

type Token string

type UpdateProgressRequest struct {
	Guarantee uint     `json:"guarantee"`
	Progress  Progress `json:"progress"`
}

type Progress struct {
	Stage    string  `json:"stage" bson:"stage"`
	Progress float32 `json:"progress" bson:"progress"`
}

type MetaProgress struct {
	Token    Token
	Progress Progress
}

type NetConfig struct {
	Address string
}

type ErrorProvider interface {
	Errors() <-chan error
}

type Service interface {
	ErrorProvider

	Listen(http, ws NetConfig) error
}

type ProgressStore interface {
	Get(token Token) (*Progress, error)
	Set(token Token, progress *Progress) error
	Delete(token Token) error
}

type Channel interface {
	ErrorProvider

	Push(MetaProgress) error
	Progresses() <-chan MetaProgress
	Close() error
}
