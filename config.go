package mongo

import (
	"reflect"
	"time"

	"github.com/go-ginger/models"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	models.IConfig

	ConnectionString string
	ReplicaSet       string
	DatabaseName     string
	CollectionNamer  func(value interface{}) string
	SetFlagOnDelete  bool
	ClientOptions    []*options.ClientOptions

	TimeoutMs int64
	Timeout   time.Duration

	MinPoolSize int
	MaxPoolSize int
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

	var timeoutMs int64
	timeout := v.FieldByName("timeout")
	if timeout.IsValid() {
		timeoutMs = timeout.Int()
	}
	if timeoutMs == 0 {
		timeoutMs = 5000
	}

	var minPoolSize int
	poolSize := v.FieldByName("minPoolSize")
	if poolSize.IsValid() {
		minPoolSize = int(poolSize.Int())
	}
	if minPoolSize == 0 {
		minPoolSize = 10
	}
	var maxPoolSize int
	poolSize = v.FieldByName("maxPoolSize")
	if poolSize.IsValid() {
		maxPoolSize = int(poolSize.Int())
	}
	if maxPoolSize == 0 {
		maxPoolSize = 50
	}

	config = Config{
		ConnectionString: connectionString.String(),
		DatabaseName:     databaseName.String(),
		CollectionNamer:  getCollectionName,
		SetFlagOnDelete:  setFlagOnDelete,
		TimeoutMs:        timeoutMs,
		MinPoolSize:      minPoolSize,
		MaxPoolSize:      maxPoolSize,
	}
	config.Timeout = time.Millisecond * time.Duration(config.TimeoutMs)
}
