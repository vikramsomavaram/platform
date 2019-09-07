/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package models

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	"github.com/tribehq/platform/lib/cache"
	"github.com/tribehq/platform/lib/database"
	"github.com/tribehq/platform/utils/webhooks"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// Page represents a page.
type Page struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt   *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy   primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Name        string             `json:"name" bson:"name"`
	Title       string             `json:"title" bson:"title"`
	Body        string             `json:"body" bson:"body"`
	Description string             `json:"description" bson:"description"`
	Language    string             `json:"language" bson:"language"`
	IsActive    bool               `json:"isActive" bson:"isActive"`
}

// CreatePage creates new page.
func CreatePage(page Page) (*Page, error) {
	page.CreatedAt = time.Now()
	page.UpdatedAt = time.Now()
	page.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(PageCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &page)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("page.created", &page)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(page.ID.Hex(), page, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &page, nil
}

// GetPageByID gives requested page by id.
func GetPageByID(ID string) (*Page, error) {
	db := database.MongoDB
	page := &Page{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(page)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	id, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		return nil, err
	}
	filter := bson.D{{"_id", id}, {"deletedAt", bson.M{"$exists": false}}}
	ctx := context.Background()
	err = db.Collection(PageCollection).FindOne(ctx, filter).Decode(&page)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, page, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return page, nil
}

// GetPages gives a list of pages.
func GetPages(filter bson.D, limit int, after *string, before *string, first *int, last *int) (pages []*Page, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(PageCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(PageCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		page := &Page{}
		err = cur.Decode(&page)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		pages = append(pages, page)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return pages, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdatePage updates page.
func UpdatePage(c *Page) (*Page, error) {
	page := c
	page.UpdatedAt = time.Now()
	filter := bson.D{{"_id", page.ID}}
	db := database.MongoDB
	pageCollection := db.Collection(PageCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := pageCollection.FindOneAndReplace(context.Background(), filter, page, findRepOpts).Decode(&page)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("page.updated", &page)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(page.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return page, nil
}

// DeletePageByID deletes page by id.
func DeletePageByID(ID string) (bool, error) {
	db := database.MongoDB
	id, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", id}}
	pageCollection := db.Collection(PageCollection)
	res, err := pageCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("page.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (page *Page) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, page); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (page *Page) MarshalBinary() ([]byte, error) {
	return json.Marshal(page)
}
