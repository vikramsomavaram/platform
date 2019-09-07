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

//ProductVariation represents product variation.
type ProductVariation struct {
	ID                primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt         time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt         *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt         time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy         primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Description       string             `json:"description" bson:"description"`
	ParentProductID   string             `json:"parentProductID" bson:"parentProductID"`
	Permalink         string             `json:"permalink" bson:"permalink"`
	Sku               string             `json:"sku" bson:"sku"`
	Price             float64            `json:"price" bson:"price"`
	RegularPrice      float64            `json:"regularPrice" bson:"regularPrice"`
	SalePrice         float64            `json:"salePrice" bson:"salePrice"`
	DateOnSaleFrom    time.Time          `json:"dateOnSaleFrom" bson:"dateOnSaleFrom"`
	DateOnSaleTo      time.Time          `json:"dateOnSaleTo" bson:"dateOnSaleTo"`
	OnSale            bool               `json:"onSale" bson:"onSale"`
	Status            ProductStatus      `json:"status" bson:"status"`
	Purchasable       bool               `json:"purchasable" bson:"purchasable"`
	Virtual           bool               `json:"virtual" bson:"virtual"`
	Downloadable      bool               `json:"downloadable" bson:"downloadable"`
	Downloads         []ProductDownload  `json:"downloads" bson:"downloads"`
	DownloadLimit     int                `json:"downloadLimit" bson:"downloadLimit"`
	DownloadExpiry    int                `json:"downloadExpiry" bson:"downloadExpiry"`
	TaxStatus         string             `json:"taxStatus" bson:"taxStatus"`
	TaxClass          string             `json:"taxClass" bson:"taxClass"`
	ManageStock       bool               `json:"manageStock" bson:"manageStock"`
	StockQuantity     int                `json:"stockQuantity" bson:"stockQuantity"`
	StockStatus       string             `json:"stockStatus" bson:"stockStatus"`
	BackOrders        string             `json:"backOrders" bson:"backOrders"`
	BackOrdersAllowed bool               `json:"backOrdersAllowed" bson:"backOrdersAllowed"`
	BackOrdered       bool               `json:"backOrdered" bson:"backOrdered"`
	Weight            float64            `json:"weight" bson:"weight"`
	Dimensions        ProductDimensions  `json:"dimensions" bson:"dimensions"`
	ShippingClass     string             `json:"shippingClass" bson:"shippingClass"`
	ShippingClassID   string             `json:"shippingClassId" bson:"shippingClassId"`
	Image             ProductImage       `json:"images" bson:"images"`
	Attributes        []ProductAttribute `json:"attributes" bson:"attributes"`
	MenuOrder         int                `json:"menuOrder" bson:"menuOrder"`
	MetaData          []ProductMetadata  `json:"metaData" bson:"metaData"`
}

// CreateProductVariation creates new product variation.
func CreateProductVariation(productVariation ProductVariation) (*ProductVariation, error) {
	productVariation.CreatedAt = time.Now()
	productVariation.UpdatedAt = time.Now()
	productVariation.ID = primitive.NewObjectID()
	db := database.MongoDB
	installationCollection := db.Collection(ProductVariationCollection)
	ctx := context.Background()
	_, err := installationCollection.InsertOne(ctx, &productVariation)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("product_variation.created", &productVariation)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(productVariation.ID.Hex(), productVariation, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &productVariation, nil
}

// GetProductVariationByID gives a product variation by id.
func GetProductVariationByID(ID primitive.ObjectID) (*ProductVariation, error) {
	db := database.MongoDB
	productVariation := &ProductVariation{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID.Hex()).Scan(productVariation)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	filter := bson.D{{"_id", ID}, {"deletedAt", bson.M{"$exists": false}}}
	ctx := context.Background()
	err = db.Collection(ProductVariationCollection).FindOne(ctx, filter).Decode(&productVariation)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID.Hex(), productVariation, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return productVariation, nil
}

// GetProductVariations gives a list of product variations.
func GetProductVariations(filter bson.D, limit int, after *string, before *string, first *int, last *int) (prodVariations []*ProductVariation, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(ProductVariationCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(ProductVariationCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		prodVariation := &ProductVariation{}
		err = cur.Decode(&prodVariation)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		prodVariations = append(prodVariations, prodVariation)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return prodVariations, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateProductVariation updates product variation.
func UpdateProductVariation(productVariation *ProductVariation) *ProductVariation {
	productVariation.UpdatedAt = time.Now()
	filter := bson.D{{"_id", productVariation.ID}}
	db := database.MongoDB
	productVariationCollection := db.Collection(ProductVariationCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := productVariationCollection.FindOneAndReplace(context.Background(), filter, productVariation, findRepOpts).Decode(&productVariation)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("product_variation.updated", &productVariation)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(productVariation.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return productVariation
}

// DeleteProductVariationByID deletes product variation by id.
func DeleteProductVariationByID(ID string) (bool, error) {
	db := database.MongoDB
	filter := bson.D{{"_id", ID}}
	productVariationCollection := db.Collection(ProductVariationCollection)
	res, err := productVariationCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	log.Warningln(res)
	if err != nil && res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("product_variation.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (productVariation *ProductVariation) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, productVariation); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (productVariation *ProductVariation) MarshalBinary() ([]byte, error) {
	return json.Marshal(productVariation)
}
