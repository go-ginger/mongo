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
	var id primitive.ObjectID
	objIDPtr, ok := req.ID.(*primitive.ObjectID)
	if ok && objIDPtr != nil {
		id = *objIDPtr
	} else {
		objID, ok := req.ID.(primitive.ObjectID)
		if ok {
			id = objID
		} else {
			id, err = primitive.ObjectIDFromHex(fmt.Sprintf("%v", req.ID))
			if err != nil {
				return errors.HandleError(err)
			}
		}
	}
	var filter *bson.M
	if req.Filters != nil {
		err = handler.NormalizeFilter(req.Filters)
		if err != nil {
			return err
		}
		var f map[string]interface{} = *req.Filters
		filter, err = getBsonDocument(&f)
	}
	if filter == nil {
		filter = &bson.M{}
	}
	if _, ok := (*filter)["_id"]; !ok {
		(*filter)["_id"] = id
	}
	improveFilter(filter, nil)
	var doc *bson.D
	var body interface{} = req.Body
	if body == nil {
		body = req.ExtraQuery
	}
	data, err := bson.Marshal(body)
	if err != nil {
		return err
	}
	err = bson.Unmarshal(data, &doc)
	var update interface{}
	if req.Body != nil {
		update = bson.M{
			"$set": doc,
		}
	} else {
		update = doc
	}
	result, err := collection.UpdateOne(*db.Context, filter, update)
	if err != nil {
		return errors.HandleError(err)
	}
	if result.MatchedCount == 0 {
		return errors.GetError(request, errors.NotFoundError)
	}
	return handler.BaseDbHandler.Update(request)
}
