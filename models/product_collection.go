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

type ProductCollection struct {
	ID            primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt     time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt     time.Time          `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt     time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy     primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Type          string             `json:"type" bson:"type"`
	IsActive      bool               `json:"isActive" bson:"isActive"`
	Name          string             `json:"name" bson:"name"`
	Description   string             `json:"description" bson:"description"`
	Slug          string             `json:"slug" bson:"slug"`
	Relationships []*ProductData     `json:"relationships" bson:"relationships"`
}

// CreateProductCollection creates new advertisement banner.
func CreateProductCollection(productCollection ProductCollection) (*ProductCollection, error) {
	productCollection.CreatedAt = time.Now()
	productCollection.UpdatedAt = time.Now()
	productCollection.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(ProductCollectionCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &productCollection)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("product_collection.created", &productCollection)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(productCollection.ID.Hex(), productCollection, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &productCollection, nil
}

// GetProductCollectionByID gets advertisement banners by ID.
func GetProductCollectionByID(ID string) *ProductCollection {
	db := database.MongoDB
	productCollection := &ProductCollection{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(productCollection)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	oID, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		return nil
	}
	filter := bson.D{{"_id", oID}, {"deletedAt", bson.M{"$exists": false}}}
	ctx := context.Background()
	err = db.Collection(ProductCollectionCollection).FindOne(ctx, filter).Decode(&productCollection)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil
		}
		return nil
	}
	//set cache item
	err = cacheClient.Set(ID, productCollection, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return productCollection
}

// GetProductCollections gets the array of advertisement banners.
func GetProductCollections(filter bson.D, limit int, after *string, before *string, first *int, last *int) (productCollections []*ProductCollection, totalCount int64, hasPrevious, hasNext bool, err error) {
	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(ProductCollectionCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(ProductCollectionCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
ctx := context.Background()
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		productCollection := &ProductCollection{}
		err = cur.Decode(&productCollection)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		productCollections = append(productCollections, productCollection)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return productCollections, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateProductCollection updates the advertisement banners.
func UpdateProductCollection(c *ProductCollection) (*ProductCollection, error) {
	productCollection := c
	productCollection.UpdatedAt = time.Now()
	filter := bson.D{{"_id", productCollection.ID}}
	db := database.MongoDB
	collection := db.Collection(ProductCollectionCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := collection.FindOneAndReplace(context.Background(), filter, productCollection, findRepOpts).Decode(&productCollection)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("product_collection.updated", &productCollection)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(productCollection.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return productCollection, nil
}

// DeleteProductCollectionByID deletes the advertisement banners by ID.
func DeleteProductCollectionByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	collection := db.Collection(ProductCollectionCollection)
	res, err := collection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	log.Warningln(res)
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("product_collection.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (productCollection *ProductCollection) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, productCollection); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (productCollection *ProductCollection) MarshalBinary() ([]byte, error) {
	return json.Marshal(productCollection)
}
