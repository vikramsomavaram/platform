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

// Restaurant represents a restaurant.
type Restaurant struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Name      string             `json:"name" bson:"name"`
	IsActive  bool               `json:"isActive" bson:"isActive"`
}

// CreateRestaurant creates new restaurant.
func CreateRestaurant(restaurant Restaurant) (*Restaurant, error) {
	restaurant.CreatedAt = time.Now()
	restaurant.UpdatedAt = time.Now()
	restaurant.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(RestaurantsCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &restaurant)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("restaurant.created", &restaurant)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(restaurant.ID.Hex(), restaurant, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &restaurant, nil
}

// GetRestaurantByID gives the requested restaurant by id.
func GetRestaurantByID(ID string) (*Restaurant, error) {
	db := database.MongoDB
	restaurant := &Restaurant{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(restaurant)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	oID, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		return nil, err
	}
	filter := bson.D{{"_id", oID}, {"deletedAt", bson.M{"$exists": false}}}
	ctx := context.Background()
	err = db.Collection(RestaurantsCollection).FindOne(ctx, filter).Decode(&restaurant)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, restaurant, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return restaurant, nil
}

// GetRestaurants gives an array of restaurants.
func GetRestaurants(filter bson.D, limit int, after *string, before *string, first *int, last *int) (restaurants []*Restaurant, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(RestaurantsCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(RestaurantsCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
ctx := context.Background()
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		restaurant := &Restaurant{}
		err = cur.Decode(&restaurant)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		restaurants = append(restaurants, restaurant)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return restaurants, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateRestaurant updates the restaurant.
func UpdateRestaurant(c *Restaurant) (*Restaurant, error) {
	restaurant := c
	restaurant.UpdatedAt = time.Now()
	filter := bson.D{{"_id", restaurant.ID}}
	db := database.MongoDB
	restaurantsCollection := db.Collection(RestaurantsCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := restaurantsCollection.FindOneAndReplace(context.Background(), filter, restaurant, findRepOpts).Decode(&restaurant)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("restaurant.updated", &restaurant)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(restaurant.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return restaurant, nil
}

// DeleteRestaurantByID deletes the restaurant by id.
func DeleteRestaurantByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	restaurantsCollection := db.Collection(RestaurantsCollection)
	res, err := restaurantsCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("restaurant.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (restaurant *Restaurant) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, restaurant); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (restaurant *Restaurant) MarshalBinary() ([]byte, error) {
	return json.Marshal(restaurant)
}
