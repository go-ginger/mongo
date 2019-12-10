package mongo

import (
	"go.mongodb.org/mongo-driver/bson"
)

func getBsonDocument(filters *map[string]interface{}) (result *bson.M, err error) {
	ds := bson.M{}
	for k, v := range *filters {
		f, ok := v.(map[string]interface{})
		if ok {
			nextedResult, e := getBsonDocument(&f)
			if e != nil {
				err = e
				return
			}
			nds := bson.M{}
			for k, v := range *nextedResult {
				nds[k] = v
			}
			v = nds
		}
		ds[k] = v
	}
	result = &ds
	return
}

func improveFilter(filter *bson.M) {
	if config.SetFlagOnDelete {
		(*filter)["deleted"] = bson.M{"$ne": true}
	}
}
