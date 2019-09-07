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

//ProductTag represents product tag.
type ProductTag struct {
	ID          primitive.ObjectID `json:"id" bson:"id"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt   time.Time          `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy   primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Name        string             `json:"name" bson:"name"`
	Slug        string             `json:"slug" bson:"slug"`
	Description string             `json:"description" bson:"description"`
	Count       int                `json:"count" bson:"count"`
	Links       Links              `json:"links" bson:"links"`
}

// CreateProductTag creates new product tag.
func CreateProductTag(productTag ProductTag) (*ProductTag, error) {
	productTag.CreatedAt = time.Now()
	productTag.UpdatedAt = time.Now()
	productTag.ID = primitive.NewObjectID()
	db := database.MongoDB
	installationCollection := db.Collection(ProductTagCollection)
	ctx := context.Background()
	_, err := installationCollection.InsertOne(ctx, &productTag)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("product_tag.created", &productTag)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(productTag.ID.Hex(), productTag, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &productTag, nil
}

// GetProductTagByID gives a product tag by id.
func GetProductTagByID(ID string) *ProductTag {
	db := database.MongoDB
	productTag := &ProductTag{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(productTag)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	oID, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		log.Errorln(err)
		return productTag
	}
	filter := bson.D{{"_id", oID}, {"deletedAt", bson.M{"$exists": false}}}
	ctx := context.Background()
	err = db.Collection(ProductTagCollection).FindOne(ctx, filter).Decode(&productTag)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil
		}
		log.Errorln(err)
		return nil
	}
	//set cache item
	err = cacheClient.Set(ID, productTag, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return productTag
}

// GetProductTags gives a list of product tags.
func GetProductTags(filter bson.D, limit int, after *string, before *string, first *int, last *int) (prodTags []*ProductTag, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(ProductTagCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(ProductTagCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		prodTag := &ProductTag{}
		err = cur.Decode(&prodTag)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		prodTags = append(prodTags, prodTag)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return prodTags, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateProductTag updates product tag.
func UpdateProductTag(p *ProductTag) *ProductTag {
	productTag := p
	productTag.UpdatedAt = time.Now()
	filter := bson.D{{"_id", productTag.ID}}
	db := database.MongoDB
	productTagsCollection := db.Collection(ProductCategoriesCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := productTagsCollection.FindOneAndReplace(context.Background(), filter, productTag, findRepOpts).Decode(&productTag)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("product_tag.updated", &productTag)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(productTag.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return productTag
}

// DeleteProductTagByID deletes product tag by id.
func DeleteProductTagByID(ID string) (bool, error) {
	db := database.MongoDB
	filter := bson.D{{"_id", ID}}
	productTagsCollection := db.Collection(ProductTagCollection)
	res, err := productTagsCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	log.Warningln(res)
	if err != nil && res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("product_tag.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (productTag *ProductTag) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, productTag); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (productTag *ProductTag) MarshalBinary() ([]byte, error) {
	return json.Marshal(productTag)
}
