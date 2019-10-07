package mongo

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func getBsonDocument(filters *map[string]interface{}) (result *bson.D, err error) {
	ds := bson.D{}
	for k, v := range *filters {
		f, ok := v.(map[string]interface{})
		if ok {
			nextedResult, e := getBsonDocument(&f)
			if e != nil {
				err = e
				return
			}
			nds := bson.D{}
			for _, d := range *nextedResult {
				nds = append(nds, d)
			}
			v = nds
		}
		ds = append(ds, primitive.E{Key: k, Value: v})
	}
	result = &ds
	return
}
