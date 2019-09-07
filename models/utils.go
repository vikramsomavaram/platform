/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package models

import (
	"context"
	"encoding/base64"
	log "github.com/sirupsen/logrus"
	"github.com/tribehq/platform/lib/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//ToCursor converts string to a opaque cursor
func ToCursor(value string) string {
	return base64.StdEncoding.EncodeToString([]byte(value))
}

//FromCursor converts opaque cursor to a string
func FromCursor(value string) (string, error) {
	cursor, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		log.Error(err)
		return "", err
	}

	return string(cursor), nil
}

type PageInfoUtility struct {
	Filter          bson.D
	Limit           int
	HasNextPage     bool
	HasPreviousPage bool
	QueryOpts       options.FindOptions
}

func calcTotalCountWithQueryFilters(collectionName string, filter bson.D, after, before *string) (tcint int, opfilter bson.D, err error) {
	db := database.MongoDB
	filter = append(filter, bson.E{"deletedAt", bson.M{"$exists": false}}) // we dont want deleted documents

	if after != nil {
		var afterID []byte
		afterID, err = base64.StdEncoding.DecodeString(*after)
		if err != nil {
			return
		}
		afterObjId, _ := primitive.ObjectIDFromHex(string(afterID))
		afterCursor := bson.E{"_id", bson.M{"$gt": afterObjId}}
		filter = append(filter, afterCursor)
	}

	if before != nil {
		var beforeID []byte
		beforeID, err = base64.StdEncoding.DecodeString(*before)
		if err != nil {
			return
		}
		beforeObjId, _ := primitive.ObjectIDFromHex(string(beforeID))
		afterCursor := bson.E{"_id", bson.M{"$lt": beforeObjId}}
		filter = append(filter, afterCursor)
	}

	var totalCount int64
	//Count total documents for pagination
	totalCount, err = db.Collection(collectionName).CountDocuments(context.Background(), filter)
	if err != nil {
		return
	}
	tcint = int(totalCount)
	opfilter = filter
	return
}

func PaginationUtility(after *string, before *string, first *int, last *int, count *int) (*PageInfoUtility, error) {
	var limit int
	var skip int
	filter := bson.D{}
	queryOpts := options.FindOptions{}
	pageInfoUtility := new(PageInfoUtility)

	if first != nil || last != nil {

		if first != nil && *count > *first {
			limit = *first
		}

		if last != nil {
			if limit != 0 && limit > *last {
				skip = limit - *last
				limit = limit - skip
			} else if limit == 0 && *count > *last {
				skip = *count - *last
			}
		}

		if skip > 0 {
			queryOpts.SetSkip(int64(skip))
		}

		if limit > 0 {
			queryOpts.SetLimit(int64(limit))
		}
	}

	if first != nil && *count > *first {
		pageInfoUtility.HasNextPage = true
	}

	if last != nil && *count > *last {
		pageInfoUtility.HasPreviousPage = true
	}

	pageInfoUtility.QueryOpts = queryOpts
	pageInfoUtility.Filter = filter
	pageInfoUtility.Limit = limit
	return pageInfoUtility, nil
}
