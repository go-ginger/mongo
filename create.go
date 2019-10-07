package mongo

import (
	"github.com/go-ginger/models"
	"github.com/go-ginger/models/errors"
)

func (handler *DbHandler) Insert(request models.IRequest) (result interface{}, err error) {
	req := request.GetBaseRequest()
	db, err := GetDb()
	if err != nil {
		return
	}
	defer func() {
		e := db.Close()
		if e != nil {
			err = e
		}
	}()
	collection := db.GetCollection(req.Body)
	res, err := collection.InsertOne(*db.Context, req.Body)
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
