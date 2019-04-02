package loadr

func NewStore(config interface{}) (ProgressStore, error) {
	switch c := config.(type) {
	case MongoConfig:
		return newMongoStore(&c)
	default:
		return &inMemory{}, nil
	}
}
