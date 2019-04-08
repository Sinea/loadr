package stores

import "github.com/Sinea/loadr/pkg/loadr"

func New(config interface{}) (loadr.Store, error) {
	switch c := config.(type) {
	case MongoConfig:
		return newMongoStore(&c)
	default:
		return newInMemoryStore()
	}
}
