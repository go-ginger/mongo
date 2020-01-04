package mongo

import (
	"fmt"
	"github.com/go-ginger/models"
	"github.com/go-ginger/models/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

func (handler *DbHandler) Delete(request models.IRequest) error {
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
	var id primitive.ObjectID
	if idPtr, ok := req.ID.(*primitive.ObjectID); ok {
		id = *idPtr
	} else {
		id, err = primitive.ObjectIDFromHex(fmt.Sprintf("%v", req.ID))
	}
	if err != nil {
		return errors.HandleError(err)
	}
	filter := bson.M{
		"_id": id,
	}
	if config.SetFlagOnDelete && *handler.SetFlagOnDelete {
		improveFilter(&filter, nil)
		update := bson.M{
			"$set": &bson.M{
				"deleted":    true,
				"deleted_at": time.Now().UTC(),
			},
		}
		result, err := collection.UpdateOne(*db.Context, filter, update)
		if err != nil {
			return errors.HandleError(err)
		}
		if result.MatchedCount == 0 {
			return errors.GetError(errors.NotFoundError)
		}
	} else {
		result, err := collection.DeleteOne(*db.Context, filter)
		if err != nil {
			return errors.HandleError(err)
		}
		if result.DeletedCount == 0 {
			return errors.GetError(errors.NotFoundError)
		}
	}
	return handler.BaseDbHandler.Delete(request)
}
