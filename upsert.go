package mongo

import (
	"github.com/go-ginger/models"
	"github.com/go-ginger/models/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (handler *DbHandler) Upsert(request models.IRequest) error {
	db, err := GetDb()
	if err != nil {
		return err
	}
	defer func() {
		e := db.Close()
		if e != nil {
			err = e
		}
	}()
	req := request.GetBaseRequest()
	model := handler.GetModelInstance()
	collection := db.GetCollection(model)
	var filter *bson.D
	if req.Filters != nil {
		var f map[string]interface{} = *req.Filters
		filter, err = getBsonDocument(&f)
	} else {
		filter = &bson.D{}
	}
	var doc *bson.D
	data, err := bson.Marshal(req.Body)
	if err != nil {
		return err
	}
	err = bson.Unmarshal(data, &doc)
	upsert := bson.D{
		bson.E{
			Key:   "$set",
			Value: doc,
		},
	}
	doUpsert := true
	result, err := collection.UpdateMany(*db.Context, filter, upsert, &options.UpdateOptions{
		Upsert: &doUpsert,
	})
	if err != nil {
		return errors.HandleError(err)
	}
	if result.MatchedCount == 0 {
		return errors.GetError(errors.NotFoundError)
	}
	return handler.BaseDbHandler.Upsert(request)
}
