package main

import (
	"github.com/Sinea/loadr/pkg/service"
	"log"
	"os"
	"strings"
	"time"
)

// TODO : Certificates for secure transport

func main() {
	channelConfig := getChannelConfig()
	channel := service.NewChannel(channelConfig)
	store := getStore()
	s := service.New(store, channel)

	backendConfig, clientsConfig := getConfigs()

	if s.Listen(backendConfig, clientsConfig) != nil {
		log.Fatal("could not start listening")
	}

	time.Sleep(time.Hour)
}

func getChannelConfig() interface{} {
	redis := os.Getenv("REDIS")
	if strings.TrimSpace(redis) != "" {
		return service.RedisConfig{Address: redis}
	}
	// Maybe rabbit? Someday...
	return nil
}

func getConfigs() (service.NetConfig, service.NetConfig) {
	backend := os.Getenv("BACKEND")
	clients := os.Getenv("CLIENTS")

	if strings.TrimSpace(backend) == "" {
		log.Fatal("invalid backend address")
	}

	if strings.TrimSpace(clients) == "" {
		log.Fatal("invalid client address")
	}

	return service.NetConfig{Address: backend}, service.NetConfig{Address: clients}
}

func getStore() service.ProgressStore {
	var config interface{} = nil

	mongo := strings.TrimSpace(os.Getenv("MONGO"))
	if mongo != "" {
		config = service.MongoConfig{
			Address:    mongo,
			User:       os.Getenv("MONGO_USER"),
			Pass:       os.Getenv("MONGO_PASS"),
			Database:   os.Getenv("MONGO_DATABASE"),
			Collection: os.Getenv("MONGO_COLLECTION"),
		}
	}

	if store, err := service.NewStore(config); err != nil {
		log.Fatal(err)
	} else {
		return store
	}

	return nil
}
