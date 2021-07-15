package mongo

import (
	"context"
	"github.com/go-ginger/models"
	"github.com/go-ginger/models/errors"
)

func (handler *DbHandler) Insert(request models.IRequest) (result models.IBaseModel, err error) {
	req := request.GetBaseRequest()
	db, err := GetDb()
	if err != nil {
		return
	}
	collection := db.GetCollection(req.Body)

	ctx, cancel := context.WithTimeout(*db.Context, config.Timeout)
	defer cancel()

	res, err := collection.InsertOne(ctx, req.Body)
	if err != nil {
		err = errors.HandleError(err)
		return
	}
	id := res.InsertedID
	req.Body.SetID(id)
	result = req.Body
	_, err = handler.BaseDbHandler.Insert(req)
	return
}
