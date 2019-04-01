package service

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type MongoConfig struct {
	Address    string
	User       string
	Pass       string
	Database   string
	Collection string
}

type mongoStore struct {
	config  MongoConfig
	session *mgo.Session
}

func (m *mongoStore) Get(token Token) (p *Progress, err error) {
	collection := m.session.DB(m.config.Database).C(m.config.Collection)
	query := collection.FindId(token)
	meta := MetaProgress{}
	err = query.One(&meta)
	p = &meta.Progress
	return
}

func (m *mongoStore) Set(token Token, progress *Progress) error {
	collection := m.session.DB(m.config.Database).C(m.config.Collection)
	_, err := collection.Upsert(bson.M{"_id": token}, bson.M{"progress": progress})

	if err != nil {
		return err
	}

	return nil
}

func (m *mongoStore) Delete(token Token) (err error) {
	collection := m.session.DB(m.config.Database).C(m.config.Collection)
	return collection.Remove(bson.M{"_id": token})
}

func newMongoStore(config MongoConfig) (ProgressStore, error) {
	address := fmt.Sprintf("mongodb://%s:%s@%s/%s", config.User, config.Pass, config.Address, config.Database)
	session, err := mgo.Dial(address)

	if err != nil {
		return nil, err
	}

	if err := session.Ping(); err != nil {
		session.Close()
		return nil, err
	}

	return &mongoStore{
		config:  config,
		session: session,
	}, nil
}
