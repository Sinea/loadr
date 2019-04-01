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

type Service interface {
	Listen(http, ws NetConfig) error
}
