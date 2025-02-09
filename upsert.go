package mongo

import (
	"context"
	"reflect"
	"strings"

	"github.com/go-ginger/models"
	"github.com/go-ginger/models/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (handler *DbHandler) getInsertOnlyFields(model interface{}) []string {
	onlyInsertFields := make([]string, 0)
	value := reflect.ValueOf(model)
	for value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	valueType := value.Type()
	for i := 0; i < value.NumField(); i++ {
		fType := valueType.Field(i)
		if fType.Type.Kind() == reflect.Struct {
			field := value.Field(i)
			nested := handler.getInsertOnlyFields(field.Interface())
			if len(nested) > 0 {
				onlyInsertFields = append(onlyInsertFields, nested...)
			}
		}
		tag, ok := fType.Tag.Lookup("mongo")
		if ok {
			tagParts := strings.Split(tag, ",")
			for _, part := range tagParts {
				if part == "insert_only" {
					bsonTag, ok := fType.Tag.Lookup("bson")
					if !ok {
						continue
					}
					bsonTagParts := strings.Split(bsonTag, ",")
					fieldName := bsonTagParts[0]
					onlyInsertFields = append(onlyInsertFields, fieldName)
					break
				}
			}
		}
	}
	return onlyInsertFields
}

func (handler *DbHandler) Upsert(request models.IRequest) error {
	db, err := GetDb()
	if err != nil {
		return err
	}
	req := request.GetBaseRequest()
	model := handler.GetModelInstance()
	collection := db.GetCollection(model)
	var filter *bson.M
	if req.Filters != nil {
		var f map[string]interface{} = *req.Filters
		filter, err = getBsonDocument(&f)
	}
	if filter == nil {
		filter = &bson.M{}
	}
	improveFilter(filter, &improveFilterOptions{
		IgnoreDeletedFilter: true,
	})
	onlyInsertFields := handler.getInsertOnlyFields(req.Body)
	doc := make([]bson.E, 0)
	setInsert := make([]bson.E, 0)
	data, err := bson.Marshal(req.Body)
	if err != nil {
		return err
	}
	err = bson.Unmarshal(data, &doc)
	if len(onlyInsertFields) > 0 {
		onlyInsertFieldsMap := map[string]bool{}
		for _, val := range onlyInsertFields {
			onlyInsertFieldsMap[val] = true
		}
		for i := len(doc) - 1; i >= 0; i-- {
			key := doc[i].Key
			if _, ok := onlyInsertFieldsMap[key]; ok {
				setInsert = append(setInsert, doc[i])
				doc = append(doc[:i], doc[i+1:]...)
			}
		}
	}
	upsert := bson.D{
		bson.E{
			Key:   "$set",
			Value: doc,
		},
	}
	if len(setInsert) > 0 {
		upsert = append(upsert,
			bson.E{
				Key:   "$setOnInsert",
				Value: setInsert,
			},
		)
	}

	ctx, cancel := context.WithTimeout(*db.Context, config.Timeout)
	defer cancel()

	doUpsert := true
	result, err := collection.UpdateMany(ctx, filter, upsert, &options.UpdateOptions{
		Upsert: &doUpsert,
	})
	if err != nil {
		return errors.HandleError(err)
	}
	if result.ModifiedCount == 0 && result.UpsertedCount == 0 {
		return errors.GetError(request, errors.NotFoundError)
	}
	return handler.BaseDbHandler.Upsert(request)
}
