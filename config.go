package mongo

import (
	"github.com/go-ginger/models"
	"reflect"
)

type Config struct {
	models.IConfig

	ConnectionString string
	ReplicaSet       string
	DatabaseName     string
	CollectionNamer  func(value interface{}) string
}

var config Config

func InitializeConfig(input interface{}) {
	v := reflect.Indirect(reflect.ValueOf(input))
	connectionString := v.FieldByName("MongoConnectionString")
	databaseName := v.FieldByName("MongoDatabaseName")

	config = Config{
		ConnectionString: connectionString.String(),
		DatabaseName:     databaseName.String(),
		CollectionNamer:  getCollectionName,
	}
}
