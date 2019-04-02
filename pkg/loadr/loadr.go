package loadr

const (
	_ uint = iota
	Storage
	Broadcast
)

type Token string

type UpdateProgressRequest struct {
	Guarantee uint     `json:"guarantee" validate:"min=0"`
	Progress  Progress `json:"progress"`
}

type Progress struct {
	Stage    string  `json:"stage" bson:"stage" validate:"min=1,max=200,regexp=^[a-zA-Z0-9]*$"`
	Progress float32 `json:"progress" bson:"progress" validate:"min=0,max:1"`
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
