package main

import (
	"log"
	"os"
	"strings"

	"github.com/Sinea/loadr/pkg/loadr"
	"github.com/Sinea/loadr/pkg/loadr/backend"
	"github.com/Sinea/loadr/pkg/loadr/channels"
	"github.com/Sinea/loadr/pkg/loadr/clients"
	"github.com/Sinea/loadr/pkg/loadr/stores"
)

func main() {
	channelConfig := getChannelConfig()
	channel := channels.New(channelConfig)
	store := getStore()
	s := loadr.New(store, channel)

	backendConfig, clientsConfig := getConfigs()

	b := backend.New(backendConfig)
	f := clients.New(clientsConfig, log.New(os.Stdout, "", 0))

	s.Run(b, f)

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
	b := os.Getenv("BACKEND")
	c := os.Getenv("CLIENTS")

	if strings.TrimSpace(b) == "" {
		log.Fatal("invalid backend address")
	}

	if strings.TrimSpace(b) == "" {
		log.Fatal("invalid client address")
	}

	return loadr.NetConfig{Address: b}, loadr.NetConfig{Address: c}
}

func getStore() loadr.Store {
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
