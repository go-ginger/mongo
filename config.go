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
	SetFlagOnDelete  bool
}

var config Config

func InitializeConfig(input interface{}) {
	v := reflect.Indirect(reflect.ValueOf(input))
	connectionString := v.FieldByName("MongoConnectionString")
	if !connectionString.IsValid() {
		panic("invalid mongo connection string")
	}
	databaseName := v.FieldByName("MongoDatabaseName")
	setFlagOnDeleteV := v.FieldByName("SetFlagOnDelete")
	setFlagOnDelete := false
	if setFlagOnDeleteV.IsValid() {
		setFlagOnDelete = setFlagOnDeleteV.Bool()
	}

	config = Config{
		ConnectionString: connectionString.String(),
		DatabaseName:     databaseName.String(),
		CollectionNamer:  getCollectionName,
		SetFlagOnDelete:  setFlagOnDelete,
	}
}
