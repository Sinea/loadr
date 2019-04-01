package main

import (
	"github.com/Sinea/loadr/pkg/loadr"
	"github.com/Sinea/loadr/pkg/service"
	"log"
	"os"
	"strings"
	"time"
)

// TODO : Certificates for secure transport

func main() {
	channelConfig := getChannelConfig()
	channel := loadr.NewChannel(channelConfig)
	store := getStore()
	s := loadr.New(store, channel)

	backendConfig, clientsConfig := getConfigs()

	if s.Listen(backendConfig, clientsConfig) != nil {
		log.Fatal("could not start listening")
	}

	time.Sleep(time.Hour)
}

func getChannelConfig() interface{} {
	redis := os.Getenv("REDIS")
	if strings.TrimSpace(redis) != "" {
		return loadr.RedisConfig{Address: redis}
	}
	// Maybe rabbit? Someday...
	return nil
}

func getConfigs() (loadr.NetConfig, loadr.NetConfig) {
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
	var config interface{} = nil

	mongo := strings.TrimSpace(os.Getenv("MONGO"))
	if mongo != "" {
		config = loadr.MongoConfig{
			Address:    mongo,
			User:       os.Getenv("MONGO_USER"),
			Pass:       os.Getenv("MONGO_PASS"),
			Database:   os.Getenv("MONGO_DATABASE"),
			Collection: os.Getenv("MONGO_COLLECTION"),
		}
	}

	if store, err := loadr.NewStore(config); err != nil {
		log.Fatal(err)
	} else {
		return store
	}

	return nil
}
