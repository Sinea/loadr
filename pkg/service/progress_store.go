package service

type ProgressStore interface {
	Get(token Token) (*Progress, error)
	Set(token Token, progress *Progress) error
	Delete(token Token) error
}

func NewStore(config interface{}) (ProgressStore, error) {
	switch c := config.(type) {
	case MongoConfig:
		return newMongoStore(c)
	default:
		return &inMemory{}, nil
	}
}
