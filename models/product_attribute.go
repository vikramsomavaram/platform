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

//ProductAttribute represents product attribute.
type ProductAttribute struct {
	ID          primitive.ObjectID `json:"id" bson:"id"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt   *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy   primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Name        string             `json:"name" bson:"name"`
	Slug        string             `json:"slug" bson:"slug"`
	Type        string             `json:"type" bson:"type"`
	OrderBy     string             `json:"orderBy" bson:"orderBy"`
	HasArchives bool               `json:"hasArchives" bson:"hasArchives"`
	Position    int                `json:"position" bson:"position"`
	Visible     bool               `json:"visible" bson:"visible"`
	Variation   bool               `json:"variation" bson:"variation"`
	Option      []string           `json:"option" bson:"option"`
}

// CreateProductAttribute creates new product attributes.
func CreateProductAttribute(product ProductAttribute) (*ProductAttribute, error) {
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()
	product.ID = primitive.NewObjectID()
	db := database.MongoDB
	installationCollection := db.Collection(productAttributeCollection)
	ctx := context.Background()
	_, err := installationCollection.InsertOne(ctx, &product)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("product_attribute.created", &product)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(product.ID.Hex(), product, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &product, nil
}

// GetProductAttributeByID gives requested product attribute by id.
func GetProductAttributeByID(ID string) (*ProductAttribute, error) {
	db := database.MongoDB
	product := &ProductAttribute{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(product)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	filter := bson.D{{"_id", ID}, {"deletedAt", bson.M{"$exists": false}}}
	ctx := context.Background()
	err = db.Collection(productAttributeCollection).FindOne(ctx, filter).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, product, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return product, nil
}

// GetProductAttributes gives a list of product attributes.
func GetProductAttributes(filter bson.D, limit int, after *string, before *string, first *int, last *int) (prodAttributes []*ProductAttribute, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(productAttributeCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(productAttributeCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		prodAttribute := &ProductAttribute{}
		err = cur.Decode(&prodAttribute)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		prodAttributes = append(prodAttributes, prodAttribute)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return prodAttributes, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateProductAttribute updates the product attribute.
func UpdateProductAttribute(product *ProductAttribute) *ProductAttribute {
	product.UpdatedAt = time.Now()
	filter := bson.D{{"_id", product.ID}}
	db := database.MongoDB
	productsCollection := db.Collection(productAttributeCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := productsCollection.FindOneAndReplace(context.Background(), filter, product, findRepOpts).Decode(&product)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("product_attribute.updated", &product)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(product.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return product
}

// DeleteProductAttributeByID deletes product attribute by id.
func DeleteProductAttributeByID(ID string) (bool, error) {
	db := database.MongoDB
	filter := bson.D{{"_id", ID}}
	productsCollection := db.Collection(productAttributeCollection)
	res, err := productsCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	log.Warningln(res)
	if err != nil && res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("product_attribute.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (product *ProductAttribute) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, product); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (product *ProductAttribute) MarshalBinary() ([]byte, error) {
	return json.Marshal(product)
}
