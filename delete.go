package mongo

import (
	"github.com/go-ginger/models"
	"github.com/go-ginger/models/errors"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go.mongodb.org/mongo-driver/bson"
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
	filter, err := getFilter(req)
	if err != nil {
		return err
	}
	if filter == nil {
		return errors.GetInternalServiceError(request,
			request.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "InvalidFilter",
					Other: "invalid filter",
				},
			}))
	}
	if config.SetFlagOnDelete && *handler.SetFlagOnDelete {
		improveFilter(filter, nil)
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
			return errors.GetError(request, errors.NotFoundError)
		}
	} else {
		result, err := collection.DeleteOne(*db.Context, filter)
		if err != nil {
			return errors.HandleError(err)
		}
		if result.DeletedCount == 0 {
			return errors.GetError(request, errors.NotFoundError)
		}
	}
	return handler.BaseDbHandler.Delete(request)
}
