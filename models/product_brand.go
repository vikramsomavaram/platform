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

type ProductBrand struct {
	ID            primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt     time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt     time.Time          `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt     time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy     primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Type          string             `json:"type" bson:"type"`
	IsActive      bool               `json:"isActive" bson:"isActive"`
	Name          string             `json:"name" bson:"name"`
	Slug          string             `json:"slug" bson:"slug"`
	Description   string             `json:"description" bson:"description"`
	Relationships []string           `json:"relationships" bson:"relationships"`
}

// CreateProductBrand creates new advertisement banner.
func CreateProductBrand(productBrand ProductBrand) (*ProductBrand, error) {
	productBrand.CreatedAt = time.Now()
	productBrand.UpdatedAt = time.Now()
	productBrand.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(ProductBrandCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &productBrand)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("product_brand.created", &productBrand)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(productBrand.ID.Hex(), productBrand, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &productBrand, nil
}

// GetProductBrandByID gets advertisement banners by ID.
func GetProductBrandByID(ID string) *ProductBrand {
	db := database.MongoDB
	productBrand := &ProductBrand{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(productBrand)
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
	err = db.Collection(ProductBrandCollection).FindOne(ctx, filter).Decode(&productBrand)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil
		}
		return nil
	}
	//set cache item
	err = cacheClient.Set(ID, productBrand, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return productBrand
}

// GetProductBrands gets the array of advertisement banners.
func GetProductBrands(filter bson.D, limit int, after *string, before *string, first *int, last *int) (productBrands []*ProductBrand, totalCount int64, hasPrevious, hasNext bool, err error) {
	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(ProductBrandCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(ProductBrandCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
ctx := context.Background()
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		productBrand := &ProductBrand{}
		err = cur.Decode(&productBrand)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		productBrands = append(productBrands, productBrand)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return productBrands, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateProductBrand updates the advertisement banners.
func UpdateProductBrand(c *ProductBrand) (*ProductBrand, error) {
	productBrand := c
	productBrand.UpdatedAt = time.Now()
	filter := bson.D{{"_id", productBrand.ID}}
	db := database.MongoDB
	collection := db.Collection(ProductBrandCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := collection.FindOneAndReplace(context.Background(), filter, productBrand, findRepOpts).Decode(&productBrand)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("product_brand.updated", &productBrand)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(productBrand.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return productBrand, nil
}

// DeleteProductBrandByID deletes the advertisement banners by ID.
func DeleteProductBrandByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	collection := db.Collection(ProductBrandCollection)
	res, err := collection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	log.Warningln(res)
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("product_brand.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (productBrand *ProductBrand) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, productBrand); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (productBrand *ProductBrand) MarshalBinary() ([]byte, error) {
	return json.Marshal(productBrand)
}
