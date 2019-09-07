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

//ProductShippingClass represents product shipping class.
type ProductShippingClass struct {
	ID          primitive.ObjectID `json:"id" json:"id"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt   time.Time          `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy   primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Name        string             `json:"name" json:"name"`
	Slug        string             `json:"slug" json:"slug"`
	Description string             `json:"description" json:"description"`
	Count       int                `json:"count" json:"count"`
	Links       Links              `json:"links" json:"links"`
}

// CreateProductShippingClass creates new product shipping class.
func CreateProductShippingClass(productShippingClass ProductShippingClass) (*ProductShippingClass, error) {
	productShippingClass.CreatedAt = time.Now()
	productShippingClass.UpdatedAt = time.Now()
	productShippingClass.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(ProductShippingClassCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &productShippingClass)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("product_shipping_class.created", &productShippingClass)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(productShippingClass.ID.Hex(), productShippingClass, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &productShippingClass, nil
}

// GetProductShippingClassByID gives a product shipping class by id.
func GetProductShippingClassByID(ID string) *ProductShippingClass {
	db := database.MongoDB
	productShippingClass := &ProductShippingClass{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(productShippingClass)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	oID, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		log.Errorln(err)
		return productShippingClass
	}
	filter := bson.D{{"_id", oID}, {"deletedAt", bson.M{"$exists": false}}}
	ctx := context.Background()
	err = db.Collection(ProductShippingClassCollection).FindOne(ctx, filter).Decode(&productShippingClass)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil
		}
		log.Errorln(err)
		return nil
	}
	//set cache item
	err = cacheClient.Set(ID, productShippingClass, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return productShippingClass
}

// GetProductShippingClasses gives a list of product shipping classes.
func GetProductShippingClasses(filter bson.D, limit int, after *string, before *string, first *int, last *int) (prodShippingClasses []*ProductShippingClass, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(ProductShippingClassCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(ProductShippingClassCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		prodShippingClass := &ProductShippingClass{}
		err = cur.Decode(&prodShippingClass)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		prodShippingClasses = append(prodShippingClasses, prodShippingClass)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return prodShippingClasses, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateProductShippingClass updates product shipping class.
func UpdateProductShippingClass(p *ProductShippingClass) *ProductShippingClass {
	productShippingClass := p
	productShippingClass.UpdatedAt = time.Now()
	filter := bson.D{{"_id", productShippingClass.ID}}
	db := database.MongoDB
	productCollection := db.Collection(ProductShippingClassCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := productCollection.FindOneAndReplace(context.Background(), filter, productShippingClass, findRepOpts).Decode(&productShippingClass)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("product_shipping_class.updated", &productShippingClass)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(productShippingClass.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return productShippingClass
}

// DeleteProductShippingClassByID deletes product shipping class by id.
func DeleteProductShippingClassByID(ID string) (bool, error) {
	db := database.MongoDB
	filter := bson.D{{"_id", ID}}
	productCollection := db.Collection(ProductShippingClassCollection)
	res, err := productCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	log.Warningln(res)
	if err != nil && res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("product_shipping_class.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (productShippingClass *ProductShippingClass) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, productShippingClass); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (productShippingClass *ProductShippingClass) MarshalBinary() ([]byte, error) {
	return json.Marshal(productShippingClass)
}
