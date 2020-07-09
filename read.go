package mongo

import (
	"fmt"
	"github.com/go-ginger/helpers"
	"github.com/go-ginger/models"
	"github.com/go-ginger/models/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"math"
)

func (handler *DbHandler) countDocuments(db *DB, collection *mongo.Collection, filter *bson.M,
	done chan bool, count *uint64, opts ...*options.CountOptions) {
	total, err := collection.CountDocuments(*db.Context, filter, opts...)
	if err != nil {
		fmt.Println(fmt.Sprintf("error on count documents: %v", err))
	}
	totalCount := uint64(total)
	*count += totalCount
	done <- true
}

func (handler *DbHandler) ImproveIDFilter(value interface{}) (result interface{}, err error) {
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
				return handler.ImproveIDFilter(filtersV)
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

func (handler *DbHandler) NormalizeFilter(filters *models.Filters) (err error) {
	if id, ok := (*filters)["id"]; ok {
		delete(*filters, "id")
		result, e := handler.ImproveIDFilter(id)
		if e != nil {
			err = e
			return
		}
		(*filters)["_id"] = result
	}
	return
}

func (handler *DbHandler) Paginate(request models.IRequest) (result *models.PaginateResult, err error) {
	db, err := GetDb()
	if err != nil {
		return nil, err
	}
	defer func() {
		e := db.Close()
		if e != nil {
			err = e
		}
	}()
	req := request.GetBaseRequest()

	var filter *bson.M
	if req.Filters != nil {
		err = handler.NormalizeFilter(req.Filters)
		if err != nil {
			return
		}
		var f map[string]interface{} = *req.Filters
		filter, err = getBsonDocument(&f)
	}
	if filter == nil {
		filter = &bson.M{}
	}
	improveFilter(filter, nil)
	offset := int64((req.Page - 1) * req.PerPage)
	limit := int64(req.PerPage)

	done := make(chan bool, 1)
	var totalCount uint64
	ms := handler.GetModelsInstance()
	collection := db.GetCollection(ms)
	go handler.countDocuments(db, collection, filter, done, &totalCount)
	findOptions := &options.FindOptions{
		Skip:  &offset,
		Limit: &limit,
	}
	if req.Sort != nil {
		sort := bson.D{}
		for _, s := range *req.Sort {
			order := 1
			if s.Ascending {
				order = 1
			} else {
				order = -1
			}
			sort = append(sort, bson.E{Key: s.Name, Value: order})
		}
		findOptions.SetSort(sort)
	}
	cur, err := collection.Find(*db.Context, *filter, findOptions)
	if err != nil {
		return
	}
	defer func() {
		e := cur.Close(*db.Context)
		if e != nil {
			err = e
		}
	}()
	queryResult := handler.GetModelsInstance()
	for cur.Next(*db.Context) {
		model := handler.GetModelInstance()
		err = cur.Decode(model)
		if err != nil {
			return
		}
		queryResult = helpers.AppendToSlice(queryResult, model)
	}
	if err = cur.Err(); err != nil {
		return
	}
	<-done
	pageCount := uint64(math.Ceil(float64(totalCount) / float64(req.PerPage)))
	return &models.PaginateResult{
		Items: queryResult,
		Pagination: models.PaginationInfo{
			Page:       req.Page,
			PerPage:    req.PerPage,
			PageCount:  pageCount,
			TotalCount: totalCount,
			HasNext:    req.Page < pageCount,
		},
	}, nil
}

func (handler *DbHandler) Get(request models.IRequest) (result models.IBaseModel, err error) {
	req := request.GetBaseRequest()
	var filter *bson.M
	if req.Filters == nil {
		req.Filters = &models.Filters{}
	}
	if req.ID != nil {
		(*req.Filters)["id"] = req.ID
	}
	err = handler.NormalizeFilter(req.Filters)
	if err != nil {
		return
	}
	var f map[string]interface{} = *req.Filters
	filter, err = getBsonDocument(&f)
	improveFilter(filter, nil)
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
	model := handler.GetModelInstance()
	collection := db.GetCollection(model)
	var limit int64 = 1
	cur, err := collection.Find(*db.Context, filter, &options.FindOptions{
		Limit: &limit,
	})
	if err != nil {
		err = errors.HandleError(err)
		return
	}
	found := false
	for cur.Next(*db.Context) {
		err = cur.Decode(model)
		if err != nil {
			return
		}
		found = true
	}
	if err = cur.Err(); err != nil {
		err = errors.HandleError(err)
		return
	}
	if !found {
		err = errors.GetNotFoundError(request)
		return
	}
	result = model.(models.IBaseModel)
	return
}
