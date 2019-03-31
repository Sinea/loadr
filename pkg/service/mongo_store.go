package service

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

type MongoConfig struct {
	Address    string
	User       string
	Pass       string
	Database   string
	Collection string
}

type mongoStore struct {
	config MongoConfig
}

func (*mongoStore) Get(token Token) (*Progress, error) {
	panic("implement me")
}

func (*mongoStore) Set(token Token, progress *Progress) error {
	panic("implement me")
}

func (*mongoStore) Delete(token Token) error {
	panic("implement me")
}

func newMongoStore(config MongoConfig) ProgressStore {

	address := fmt.Sprintf("mongodb://%s", config.Address)
	client, err := mongo.Connect(context.Background(), &options.ClientOptions{
		Hosts: []string{address},
	})

	if err != nil {
		log.Fatal(err)

	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	return &mongoStore{
		config: config,
	}
}
