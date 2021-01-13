package mongo

import (
	"fmt"
	"github.com/go-ginger/models"
	"github.com/go-ginger/models/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

type improveFilterOptions struct {
	IgnoreDeletedFilter bool
}

func improveFilter(filter *bson.M, deketeOptions *improveFilterOptions) {
	if deketeOptions == nil || !deketeOptions.IgnoreDeletedFilter {
		if config.SetFlagOnDelete {
			(*filter)["deleted"] = bson.M{"$ne": true}
		}
	}
}

func improveIDFilter(value interface{}) (result interface{}, err error) {
	result = value
	filters, ok := value.(*models.Filters)
	if !ok {
		var filtersMap map[string]interface{}
		filtersMap, ok = value.(map[string]interface{})
		if ok {
			f := models.Filters(filtersMap)
			filters = &f
		}
	}
	if ok {
		for k, v := range *filters {
			strV, ok := v.(string)
			if ok {
				(*filters)[k], err = primitive.ObjectIDFromHex(fmt.Sprintf("%v", strV))
				if err != nil {
					return
				}
				return
			}
			strsV, ok := v.([]string)
			if ok {
				ids := make([]primitive.ObjectID, 0)
				for _, str := range strsV {
					id, e := primitive.ObjectIDFromHex(fmt.Sprintf("%v", str))
					if e != nil {
						err = e
						return
					}
					ids = append(ids, id)
				}
				(*filters)[k] = ids
				return
			}
			filtersV, ok := v.(*models.Filters)
			if ok {
				return improveIDFilter(filtersV)
			}
		}
		return
	}

	var strId string
	if strValue, ok := value.(string); ok {
		strId = fmt.Sprintf("%v", strValue)
	} else if strValue, ok := value.(*string); ok {
		strId = fmt.Sprintf("%v", *strValue)
	}
	if strId != "" {
		result, err = primitive.ObjectIDFromHex(strId)
	}
	return
}

func normalizeFilter(filters *models.Filters) (err error) {
	if id, ok := (*filters)["id"]; ok {
		delete(*filters, "id")
		result, e := improveIDFilter(id)
		if e != nil {
			err = e
			return
		}
		(*filters)["_id"] = result
	}
	return
}

func getFilter(request models.IRequest) (filter *bson.M, err error) {
	req := request.GetBaseRequest()
	if req.Filters != nil {
		err := normalizeFilter(req.Filters)
		if err != nil {
			return nil, err
		}
		var f map[string]interface{} = *req.Filters
		filter, err = getBsonDocument(&f)
	}
	var id *primitive.ObjectID
	if req.ID != nil {
		if idPtr, ok := req.ID.(*primitive.ObjectID); ok {
			id = idPtr
		} else {
			_id, err := primitive.ObjectIDFromHex(fmt.Sprintf("%v", req.ID))
			if err != nil {
				return nil, errors.HandleError(err)
			}
			id = &_id
		}
		if id != nil {
			if filter == nil {
				filter = &bson.M{}
			}
			(*filter)["_id"] = id
		}
	}
	return filter, nil
}
