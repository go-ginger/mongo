package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/event"
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
	collection := db.Client.
		Database(config.DatabaseName).
		Collection(config.CollectionNamer(model))
	return collection
}

func GetDb() (db *DB, err error) {
	if currentDb != nil {
		return currentDb, nil
	}
	monitor := &event.PoolMonitor{
		Event: HandlePoolMonitor,
	}
	opts := []*options.ClientOptions{
		options.Client().
			ApplyURI(config.ConnectionString).
			SetMinPoolSize(uint64(config.MinPoolSize)).
			SetMaxPoolSize(uint64(config.MaxPoolSize)).
			SetHeartbeatInterval(time.Second * 5).
			SetPoolMonitor(monitor),
	}
	opts = append(opts, config.ClientOptions...)
	ctx := context.Background()
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

func HandlePoolMonitor(evt *event.PoolEvent) {
	switch evt.Type {
	case event.PoolClosedEvent:
		reconnect()
	}
}

func reconnect() {
	for {
		if _, err := GetDb(); err == nil {
			break
		}
		time.Sleep(config.Timeout)
	}
}
