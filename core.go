package mongo

import (
	"bytes"
	"reflect"
	"strings"

	"github.com/go-ginger/dl"
	"github.com/jinzhu/inflection"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type DbHandler struct {
	dl.BaseDbHandler
	pool *MongoPool
}

func (handler *DbHandler) Initialize(h dl.IBaseDbHandler, model interface{}) {
	handler.pool = &MongoPool{
		pool:        make(chan *mongo.Client, config.MaxPoolSize),
		connections: 0,
		timeout:     config.Timeout,
		uri:         config.ConnectionString,
		poolSize:    int(config.MinPoolSize),
	}
	h.GetBaseDbHandler().Initialize(h, model)
}

func (handler *DbHandler) IdEquals(id1 interface{}, id2 interface{}) bool {
	if objId, ok := id1.(primitive.ObjectID); ok {
		id1 = objId.Hex()
	} else if objId, ok := id1.(*primitive.ObjectID); ok {
		id1 = objId.Hex()
	}
	if objId, ok := id2.(primitive.ObjectID); ok {
		id2 = objId.Hex()
	} else if objId, ok := id2.(*primitive.ObjectID); ok {
		id2 = objId.Hex()
	}
	return id1 == id2
}

type iCustomName interface {
	Name() string
}

var sMap = newSafeMap()

func getCollectionName(value interface{}) string {
	if cn, ok := value.(iCustomName); ok {
		name := cn.Name()
		if name != "" {
			return name
		}
	}
	reflectType := reflect.ValueOf(value).Type()
	for reflectType.Kind() == reflect.Slice || reflectType.Kind() == reflect.Ptr {
		reflectType = reflectType.Elem()
	}
	return inflection.Plural(defaultNamer(reflectType.Name()))
}

func defaultNamer(name string) string {
	const (
		lower = false
		upper = true
	)

	if v := sMap.Get(name); v != "" {
		return v
	}

	if name == "" {
		return ""
	}

	var (
		value                                    = commonInitialismsReplacer.Replace(name)
		buf                                      = bytes.NewBufferString("")
		lastCase, currCase, nextCase, nextNumber bool
	)

	for i, v := range value[:len(value)-1] {
		nextCase = value[i+1] >= 'A' && value[i+1] <= 'Z'
		nextNumber = value[i+1] >= '0' && value[i+1] <= '9'

		if i > 0 {
			if currCase == upper {
				if lastCase == upper && (nextCase == upper || nextNumber == upper) {
					buf.WriteRune(v)
				} else {
					if value[i-1] != '_' && value[i+1] != '_' {
						buf.WriteRune('_')
					}
					buf.WriteRune(v)
				}
			} else {
				buf.WriteRune(v)
				if i == len(value)-2 && (nextCase == upper && nextNumber == lower) {
					buf.WriteRune('_')
				}
			}
		} else {
			currCase = upper
			buf.WriteRune(v)
		}
		lastCase = currCase
		currCase = nextCase
	}

	buf.WriteByte(value[len(value)-1])

	s := strings.ToLower(buf.String())
	sMap.Set(name, s)
	return s
}
