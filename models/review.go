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

// StoreReview represents a store review.
type StoreReview struct {
	ID         primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt  time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt  *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt  time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy  primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Restaurant RestaurantReview   `json:"restaurant" bson:"restaurant"`
	Providers  ProviderReview     `json:"providers" bson:"providers"`
	Users      UserReview         `json:"users" bson:"users"`
	IsActive   bool               `json:"isActive" bson:"isActive"`
}

// CreateStoreReview creates store review.
func CreateStoreReview(storeReview StoreReview) (*StoreReview, error) {
	storeReview.CreatedAt = time.Now()
	storeReview.UpdatedAt = time.Now()
	storeReview.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(StoreReviewsCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &storeReview)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("store_review.created", &storeReview)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(storeReview.ID.Hex(), storeReview, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &storeReview, nil
}

// GetStoreReviewByID gives a store review by id.
func GetStoreReviewByID(ID string) (*StoreReview, error) {
	db := database.MongoDB
	storeReview := &StoreReview{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(storeReview)
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
	err = db.Collection(StoreReviewsCollection).FindOne(ctx, filter).Decode(&storeReview)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, storeReview, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return storeReview, nil
}

// GetStoreReviews gives a list of store reviews.
func GetStoreReviews(filter bson.D, limit int, after *string, before *string, first *int, last *int) (storeReviews []*StoreReview, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(StoreReviewsCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(StoreReviewsCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
ctx := context.Background()
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		storeReview := &StoreReview{}
		err = cur.Decode(&storeReview)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		storeReviews = append(storeReviews, storeReview)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return storeReviews, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateStoreReview updates store reviews.
func UpdateStoreReview(c *StoreReview) (*StoreReview, error) {
	storeReview := c
	storeReview.UpdatedAt = time.Now()
	filter := bson.D{{"_id", storeReview.ID}}
	db := database.MongoDB
	storeReviewsCollection := db.Collection(StoreReviewsCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := storeReviewsCollection.FindOneAndReplace(context.Background(), filter, storeReview, findRepOpts).Decode(&storeReview)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("store_review.updated", &storeReview)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(storeReview.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return storeReview, nil
}

// DeleteStoreReviewByID deletes store reviews by id.
func DeleteStoreReviewByID(ID string) (bool, error) {
	db := database.MongoDB
	id, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", id}}
	storeReviewsCollection := db.Collection(StoreReviewsCollection)
	res, err := storeReviewsCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("store_review.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (storeReview *StoreReview) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, storeReview); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (storeReview *StoreReview) MarshalBinary() ([]byte, error) {
	return json.Marshal(storeReview)
}

// Review represents a review.
type Review struct {
	ID                    primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt             time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt             *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt             time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy             primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	JobID                 primitive.ObjectID `json:"jobNumber" bson:"jobNumber"`
	UserID                string             `json:"userID" bson:"userID"`
	UserAverageRating     float64            `json:"userAverageRating" bson:"userAverageRating"`
	ProviderAverageRating float64            `json:"providerAverageRating" bson:"providerAverageRating"`
	ProviderID            string             `json:"providerID" bson:"providerID"`
	UserName              string             `json:"userName" bson:"userName"`
	ProviderName          string             `json:"providerName" bson:"providerName"`
	UserRating            float64            `json:"userRating" bson:"userRating"`
	ProviderRating        float64            `json:"providerRating" bson:"providerRating"`
	Type                  ReviewType         `json:"type" bson:"type"` //ride , store , service , delivery
	From                  string             `json:"from" bson:"from"`
	To                    string             `json:"to" bson:"to"`
	Date                  time.Time          `json:"date" bson:"date"`
	Comment               string             `json:"comment" bson:"comment"`
	IsActive              bool               `json:"isActive" bson:"isActive"`
}

// CreateReview creates new reviews.
func CreateReview(review Review) (*Review, error) {
	review.CreatedAt = time.Now()
	review.UpdatedAt = time.Now()
	review.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(ReviewsCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &review)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("review.created", &review)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(review.ID.Hex(), review, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &review, nil
}

// GetReviewByID gives review by id.
func GetReviewByID(ID string) (*Review, error) {
	db := database.MongoDB
	review := &Review{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(review)
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
	err = db.Collection(ReviewsCollection).FindOne(ctx, filter).Decode(&review)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, review, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return review, nil
}

// GetReviews gives a list of reviews.
func GetReviews(filter bson.D, limit int, after *string, before *string, first *int, last *int) (reviews []*Review, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(ReviewsCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(ReviewsCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
ctx := context.Background()
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		review := &Review{}
		err = cur.Decode(&review)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		reviews = append(reviews, review)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return reviews, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateReview updates reviews.
func UpdateReview(c *Review) (*Review, error) {
	review := c
	review.UpdatedAt = time.Now()
	filter := bson.D{{"_id", review.ID}}
	db := database.MongoDB
	reviewsCollection := db.Collection(ReviewsCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := reviewsCollection.FindOneAndReplace(context.Background(), filter, review, findRepOpts).Decode(&review)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("review.updated", &review)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(review.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return review, nil
}

// DeleteReviewByID deletes reviews by id.
func DeleteReviewByID(ID string) (bool, error) {
	db := database.MongoDB
	id, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", id}}
	reviewsCollection := db.Collection(ReviewsCollection)
	res, err := reviewsCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("review.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (review *Review) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, review); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (review *Review) MarshalBinary() ([]byte, error) {
	return json.Marshal(review)
}
