package mongo

import (
	"fmt"
	"github.com/go-ginger/models"
	"github.com/go-ginger/models/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"math"
)

func (handler *DbHandler) countDocuments(db *DB, collection *mongo.Collection, filter *bson.D,
	done chan bool, count *uint64, opts ...*options.CountOptions) {
	total, err := collection.CountDocuments(*db.Context, filter, opts...)
	if err != nil {
		fmt.Println(fmt.Sprintf("error on count documents: %v", err))
	}
	totalCount := uint64(total)
	*count += totalCount
	done <- true
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

	var filter *bson.D
	if req.Filters != nil {
		var f map[string]interface{} = *req.Filters
		filter, err = getBsonDocument(&f)
	} else {
		filter = &bson.D{}
	}
	offset := int64((req.Page - 1) * req.PerPage)
	limit := int64(req.PerPage)

	done := make(chan bool, 1)
	var totalCount uint64
	collection := db.GetCollection(req.Models)
	go handler.countDocuments(db, collection, filter, done, &totalCount)
	cur, err := collection.Find(*db.Context, *filter,
		&options.FindOptions{
			Skip:  &offset,
			Limit: &limit,
		})
	if err != nil {
		return
	}
	defer func() {
		e := cur.Close(*db.Context)
		if e != nil {
			err = e
		}
	}()
	queryResult := make([]interface{}, 0)
	for cur.Next(*db.Context) {
		err = cur.Decode(req.Model)
		if err != nil {
			return
		}
		queryResult = append(queryResult, req.Model)
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

	var filter *bson.D
	if req.Filters != nil {
		var f map[string]interface{} = *req.Filters
		if id, ok := f["id"]; ok {
			delete(f, "id")
			_id, err := primitive.ObjectIDFromHex(fmt.Sprintf("%v", id))
			if err != nil {
				return nil, errors.HandleError(err)
			}
			f["_id"] = _id
		}
		filter, err = getBsonDocument(&f)
	} else {
		filter = &bson.D{}
	}
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
	collection := db.GetCollection(req.Model)
	err = collection.FindOne(*db.Context, filter).Decode(req.Model)
	if err != nil {
		err = errors.HandleError(err)
		return
	}

	return req.Model, nil
}
