package main

import (
	"log"
	"os"
	"strings"

	"github.com/Sinea/loadr/pkg/loadr"
	"github.com/Sinea/loadr/pkg/loadr/channels"
	"github.com/Sinea/loadr/pkg/loadr/stores"
)

func main() {
	channelConfig := getChannelConfig()
	channel := channels.New(channelConfig)
	store := getStore()
	s := loadr.New(store, channel)

	backendConfig, clientsConfig := getConfigs()

	if s.Listen(backendConfig, clientsConfig) != nil {
		log.Fatal("could not start listening")
	}

	for {
		log.Println(<-s.Errors())
	}
}

func getChannelConfig() interface{} {
	redis := os.Getenv("REDIS")
	if strings.TrimSpace(redis) != "" {
		return channels.RedisConfig{Address: redis}
	}
	// Maybe rabbit? Someday...
	return nil
}

func getConfigs() (backendCfg, frontendCfg loadr.NetConfig) {
	backend := os.Getenv("BACKEND")
	clients := os.Getenv("CLIENTS")

	if strings.TrimSpace(backend) == "" {
		log.Fatal("invalid backend address")
	}

	if strings.TrimSpace(clients) == "" {
		log.Fatal("invalid client address")
	}

	return loadr.NetConfig{Address: backend}, loadr.NetConfig{Address: clients}
}

func getStore() loadr.ProgressStore {
	var config interface{}

	mongo := strings.TrimSpace(os.Getenv("MONGO"))
	if mongo != "" {
		config = stores.MongoConfig{
			Address:    mongo,
			User:       os.Getenv("MONGO_USER"),
			Pass:       os.Getenv("MONGO_PASS"),
			Database:   os.Getenv("MONGO_DATABASE"),
			Collection: os.Getenv("MONGO_COLLECTION"),
		}
	}

	if store, err := stores.New(config); err != nil {
		log.Fatalf("error creating store: %s", err)
	} else {
		return store
	}

	return nil
}
