package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

type DB struct {
	Client  *mongo.Client
	Context context.Context
}

func (db *DB) GetCollection(model interface{}) *mongo.Collection {
	collection := db.Client.
		Database(config.DatabaseName).
		Collection(config.CollectionNamer(model))
	return collection
}

func (handler *DbHandler) GetDb() (db *DB, err error) {
	ctx := context.Background()
	conn, err := handler.pool.GetConnection()
	if err != nil {
		return nil, err
	}
	return &DB{conn, ctx}, nil
}
