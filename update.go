package mongo

import (
	"encoding/json"
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
	collection := db.GetCollection(req.Model)
	id, err := primitive.ObjectIDFromHex(fmt.Sprintf("%v", req.ID))
	if err != nil {
		return errors.HandleError(err)
	}
	filter := bson.D{
		{Key: "_id", Value: id},
	}
	bytes, err := json.Marshal(req.Body)
	if err != nil {
		return err
	}
	var m map[string]interface{}
	err = json.Unmarshal(bytes, &m)
	if err != nil {
		return err
	}
	update := map[string]interface{}{
		"$set": m,
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
