package service

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	client *mongo.Client
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

func newMongoStore(config MongoConfig) (ProgressStore, error) {

	address := fmt.Sprintf("mongodb://%s", config.Address)
	client, err := mongo.Connect(context.Background(), &options.ClientOptions{
		Hosts: []string{address},
	})

	if err != nil {
		return nil, err
	}

	// Check the connection
	err = client.Ping(context.Background(), nil)

	if err != nil {
		return nil, err
	}

	fmt.Println("Connected to MongoDB!")

	return &mongoStore{
		config: config,
		client: client,
	}, nil
}
