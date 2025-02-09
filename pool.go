package mongo

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoPool struct {
	pool        chan *mongo.Client
	timeout     time.Duration
	uri         string
	connections int
	poolSize    int
}

func (mp *MongoPool) getContextTimeOut() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), mp.timeout)
	return ctx
}

func (mp *MongoPool) createToChan() {
	var conn *mongo.Client
	conn, e := mongo.NewClient(options.Client().ApplyURI(mp.uri))
	if e != nil {
		log.Printf("Create the Pool failed, err=%v", e)
	}
	e = conn.Connect(mp.getContextTimeOut())
	if e != nil {
		log.Printf("Create the Pool failed, err=%v", e)
	}
	mp.pool <- conn
	mp.connections++
}

func (mp *MongoPool) CloseConnection(conn *mongo.Client) error {
	select {
	case mp.pool <- conn:
		return nil
	default:
		if err := conn.Disconnect(context.TODO()); err != nil {
			log.Printf("Close the Pool failed, err=%v", err)
			return err
		}
		mp.connections--
		return nil
	}
}

func (mp *MongoPool) GetConnection() (*mongo.Client, error) {
	for {
		select {
		case conn := <-mp.pool:
			err := conn.Ping(mp.getContextTimeOut(), readpref.Primary())
			if err != nil {
				log.Printf("error on ping pool connection, err=%v", err)
				return nil, err
			}
			return conn, nil
		default:
			if mp.connections < mp.poolSize {
				mp.createToChan()
			}
		}
	}
}

func GetCollection(conn *mongo.Client, dbname, collection string) *mongo.Collection {
	return conn.Database(dbname).Collection(collection)
}
