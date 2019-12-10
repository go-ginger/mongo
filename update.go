package mongo

import (
	"fmt"
	"github.com/go-ginger/models"
	"github.com/go-ginger/models/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (handler *DbHandler) Update(request models.IRequest) error {
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
	id, err := primitive.ObjectIDFromHex(fmt.Sprintf("%v", req.ID))
	if err != nil {
		return errors.HandleError(err)
	}
	filter := bson.M{
		"_id": id,
	}
	improveFilter(&filter)
	var doc *bson.D
	data, err := bson.Marshal(req.Body)
	if err != nil {
		return err
	}
	err = bson.Unmarshal(data, &doc)
	update := bson.M{
		"$set": doc,
	}
	result, err := collection.UpdateOne(*db.Context, filter, update)
	if err != nil {
		return errors.HandleError(err)
	}
	if result.MatchedCount == 0 {
		return errors.GetError(errors.NotFoundError)
	}
	return handler.BaseDbHandler.Update(request)
}
