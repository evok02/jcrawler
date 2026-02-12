package db

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type Storage struct {
	DB  *mongo.Client
	ctx context.Context
}

type Connecter interface {
	Connect(opts ...*options.ClientOptions)
}

func NewStorage(uri string) (*Storage, error) {
	client, err := connect(uri)
	if err != nil {
		return nil, fmt.Errorf("NewStorage: %s", err.Error())
	}
	return &Storage{
		DB:  client,
		ctx: context.Background(),
	}, nil
}

func connect(uri string) (*mongo.Client, error) {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, fmt.Errorf("connect: %s", err.Error())
	}
	return client, nil
}

func (s *Storage) Init() error {
	var result bson.M
	context, cancel := context.WithTimeout(s.ctx, time.Second)
	defer cancel()
	if err := s.DB.Database("admin").RunCommand(context, bson.D{{Key: "ping", Value: 1}}).Decode(&result); err != nil {
		return fmt.Errorf("Init: %s", err.Error())
	}
	return s.CreateCollection("pages")
}

func (s *Storage) CreateCollection(name string) error {
	context, cancel := context.WithTimeout(s.ctx, time.Second)
	defer cancel()
	return s.DB.Database("crawler").CreateCollection(context, name)
}

func (s *Storage) CloseConnection() error {
	context, cancel := context.WithTimeout(s.ctx, time.Second)
	defer cancel()
	return s.DB.Disconnect(context)
}
