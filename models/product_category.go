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

// ProductCategory represents a product category.
type ProductCategory struct {
	ID           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt    time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt    time.Time          `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt    time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy    primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Name         string             `json:"name" bson:"name"`
	Slug         string             `json:"slug" bson:"slug"`
	Parent       int                `json:"parent" bson:"parent"`
	Store        string             `json:"store" bson:"store"`
	Description  string             `json:"description" bson:"description"`
	DisplayOrder int                `json:"displayOrder" bson:"displayOrder"`
	Display      string             `json:"display" bson:"display"`
	Image        ProductImage       `json:"image" bson:"image"`
	ServiceType  StoreCategory      `json:"serviceType" bson:"serviceType"`
	MenuCategory string             `json:"menuCategory" bson:"menuCategory"`
	MenuOrder    int                `json:"menuOrder" bson:"menuOrder"`
	Count        int                `json:"count" bson:"count"`
	IsActive     bool               `json:"isActive" bson:"isActive"`
}

//Self represents self.
type Self struct {
	Href string `json:"href" bson:"href"`
}

//Collection represents collection.
type Collection struct {
	Href string `json:"href" bson:"href"`
}

//Links represents links.
type Links struct {
	Self       []Self       `json:"self" bson:"self"`
	Collection []Collection `json:"collection" bson:"collection"`
}

//UnmarshalBinary required for the redis cache to work
func (productCategory *ProductCategory) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, productCategory); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (productCategory *ProductCategory) MarshalBinary() ([]byte, error) {
	return json.Marshal(productCategory)
}

// CreateProductCategory creates new product category.
func CreateProductCategory(productCategory ProductCategory) (*ProductCategory, error) {
	productCategory.CreatedAt = time.Now()
	productCategory.UpdatedAt = time.Now()
	productCategory.ID = primitive.NewObjectID()
	db := database.MongoDB
	installationCollection := db.Collection(ProductCategoriesCollection)
	ctx := context.Background()
	_, err := installationCollection.InsertOne(ctx, &productCategory)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("product_category.created", &productCategory)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(productCategory.ID.Hex(), productCategory, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &productCategory, nil
}

// GetProductCategoryByID gives a product category by id.
func GetProductCategoryByID(ID string) *ProductCategory {
	db := database.MongoDB
	productCategory := &ProductCategory{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(productCategory)
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
	err = db.Collection(ProductCategoriesCollection).FindOne(ctx, filter).Decode(&productCategory)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil
		}
		log.Errorln(err)
		return nil
	}
	//set cache item
	err = cacheClient.Set(ID, productCategory, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return productCategory
}

// GetProductCategories gives a list of product actegories.
func GetProductCategories(filter bson.D, limit int, after *string, before *string, first *int, last *int) (prodCategories []*ProductCategory, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(ProductCategoriesCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(ProductCategoriesCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
ctx := context.Background()
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		prodCategory := &ProductCategory{}
		err = cur.Decode(&prodCategory)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		prodCategories = append(prodCategories, prodCategory)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return prodCategories, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateProductCategory updates product category.
func UpdateProductCategory(p *ProductCategory) (*ProductCategory, error) {
	productCategory := p
	productCategory.UpdatedAt = time.Now()
	filter := bson.D{{"_id", productCategory.ID}}
	db := database.MongoDB
	productCategoriesCollection := db.Collection(ProductCategoriesCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := productCategoriesCollection.FindOneAndReplace(context.Background(), filter, productCategory, findRepOpts).Decode(&productCategory)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("product_category.updated", &productCategory)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(productCategory.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return productCategory, nil
}

// DeleteProductCategoryByID deletes product category by id.
func DeleteProductCategoryByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		return false, err
	}
	filter := bson.D{{"_id", oID}}
	productCategoriesCollection := db.Collection(ProductCategoriesCollection)
	res, err := productCategoriesCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("product_category.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}
