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

//ProductReview represents product review.
type ProductReview struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id"`
	CreatedAt          time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt          *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt          time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy          primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	ProductID          primitive.ObjectID `json:"productID" bson:"productID"`
	Status             string             `json:"status" bson:"status"`
	Reviewer           string             `json:"reviewer" bson:"reviewer"`
	ReviewerEmail      string             `json:"reviewerEmail" bson:"reviewerEmail"`
	Review             string             `json:"review" bson:"review"`
	Rating             int                `json:"rating" bson:"rating"`
	Verified           bool               `json:"verified" bson:"verified"`
	ReviewerAvatarURLs ReviewerAvatarUrls `json:"reviewerAvatarURLs" bson:"reviewerAvatarURLs"`
}

// CreateProductReview creates new product review.
func CreateProductReview(productReview ProductReview) (*ProductReview, error) {
	productReview.CreatedAt = time.Now()
	productReview.UpdatedAt = time.Now()
	productReview.ID = primitive.NewObjectID()
	db := database.MongoDB
	installationCollection := db.Collection(ProductReviewCollection)
	ctx := context.Background()
	_, err := installationCollection.InsertOne(ctx, &productReview)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("product_review.created", &productReview)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(productReview.ID.Hex(), productReview, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &productReview, nil
}

// GetProductReviewByID gives a product review by id.
func GetProductReviewByID(ID string) *ProductReview {
	db := database.MongoDB
	productReview := &ProductReview{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(productReview)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	oID, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		log.Errorln(err)
		return productReview
	}
	filter := bson.D{{"_id", oID}, {"deletedAt", bson.M{"$exists": false}}}
	ctx := context.Background()
	err = db.Collection(ProductReviewCollection).FindOne(ctx, filter).Decode(&productReview)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil
		}
		log.Errorln(err)
		return nil
	}
	//set cache item
	err = cacheClient.Set(ID, productReview, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return productReview
}

// GetProductReviews gives a list of product reviews.
func GetProductReviews(filter bson.D, limit int, after *string, before *string, first *int, last *int) (prodReviews []*ProductReview, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(ProductReviewCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(ProductReviewCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
ctx := context.Background()
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		prodReview := &ProductReview{}
		err = cur.Decode(&prodReview)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		prodReviews = append(prodReviews, prodReview)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return prodReviews, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateProductReview updates product review.
func UpdateProductReview(productReview *ProductReview) *ProductReview {
	productReview.UpdatedAt = time.Now()
	filter := bson.D{{"_id", productReview.ID}}
	db := database.MongoDB
	productReviewCollection := db.Collection(ProductReviewCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := productReviewCollection.FindOneAndReplace(context.Background(), filter, productReview, findRepOpts).Decode(&productReview)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("product_review.updated", &productReview)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(productReview.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return productReview
}

// DeleteProductReviewByID deletes product review by id.
func DeleteProductReviewByID(ID string) (bool, error) {
	db := database.MongoDB
	filter := bson.D{{"_id", ID}}
	productReviewCollection := db.Collection(ProductReviewCollection)
	res, err := productReviewCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	log.Warningln(res)
	if err != nil && res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("product_review.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (productReview *ProductReview) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, productReview); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (productReview *ProductReview) MarshalBinary() ([]byte, error) {
	return json.Marshal(productReview)
}
