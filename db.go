package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	ctx := context.TODO()
	client, err := mongo.Connect(ctx, opts...)
	if err != nil {
		return
	}
	currentDb = &DB{
		Client:  client,
		Context: &ctx,
	}
	return currentDb, nil
}
