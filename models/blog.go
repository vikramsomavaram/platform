/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package models

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/tribehq/platform/lib/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Blog struct {
}

type BlogPost struct {
}

func GetBlogPosts(filter bson.D, limit int, after *string, before *string, first *int, last *int) (addresses []*Address, totalCount int64, hasPrevious, hasNext bool, err error) {
	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(BlogPostsCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(BlogPostsCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
ctx := context.Background()
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		address := &Address{}
		err = cur.Decode(&address)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		addresses = append(addresses, address)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return addresses, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}
