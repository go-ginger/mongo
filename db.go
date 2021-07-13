package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type DB struct {
	Client  *mongo.Client
	Context *context.Context
}

var currentDb *DB

func (db *DB) Close() (err error) {
	if db.Context == nil {
		// database is not connected yet
		return
	}
	return db.Client.Disconnect(*db.Context)
}

func (db *DB) GetCollection(model interface{}) *mongo.Collection {
	collection := db.Client.Database(config.DatabaseName).Collection(config.CollectionNamer(model))
	return collection
}

func GetDb() (db *DB, err error) {
	if currentDb != nil {
		return currentDb, nil
	}
	opts := []*options.ClientOptions{
		options.Client().ApplyURI(config.ConnectionString),
	}
	opts = append(opts, config.ClientOptions...)
	client, err := mongo.Connect(context.Background(), opts...)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		return
	}
	currentDb = &DB{
		Client:  client,
		Context: &ctx,
	}
	return currentDb, nil
}
