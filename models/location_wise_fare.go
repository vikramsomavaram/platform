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

// LocationWiseFare represents a location wise fare.
type LocationWiseFare struct {
	ID                  primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt           time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt           *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt           time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy           primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	SourceLocation      string             `json:"sourceLocation" bson:"sourceLocation"`
	DestinationLocation string             `json:"destinationLocation" bson:"destinationLocation"`
	FlatFare            string             `json:"flatFare" bson:"flatFare"`
	VehicleType         string             `json:"vehicleType" bson:"vehicleType"`
	IsActive            bool               `json:"isActive" bson:"isActive"`
}

// CreateLocationWiseFare creates location wise fare.
func CreateLocationWiseFare(fare LocationWiseFare) (*LocationWiseFare, error) {
	fare.CreatedAt = time.Now()
	fare.UpdatedAt = time.Now()
	fare.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(LocationWiseFareCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &fare)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("location_wise_fare.created", &fare)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(fare.ID.Hex(), fare, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &fare, nil
}

// GetLocationWiseFareByID gives requested location wise fare by id.
func GetLocationWiseFareByID(ID string) (*LocationWiseFare, error) {
	db := database.MongoDB
	fare := &LocationWiseFare{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(fare)
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
	err = db.Collection(LocationWiseFareCollection).FindOne(ctx, filter).Decode(&fare)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, fare, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return fare, nil
}

// GetLocationWiseFares gives a list of location wise fares.
func GetLocationWiseFares(filter bson.D, limit int, after *string, before *string, first *int, last *int) (locationWiseFares []*LocationWiseFare, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(LocationWiseFareCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(LocationWiseFareCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		locationWiseFare := &LocationWiseFare{}
		err = cur.Decode(&locationWiseFare)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		locationWiseFares = append(locationWiseFares, locationWiseFare)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return locationWiseFares, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateLocationWiseFare updates location wise fare.
func UpdateLocationWiseFare(c *LocationWiseFare) (*LocationWiseFare, error) {
	fare := c
	fare.UpdatedAt = time.Now()
	filter := bson.D{{"_id", fare.ID}}
	db := database.MongoDB
	Collection := db.Collection(LocationWiseFareCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := Collection.FindOneAndReplace(context.Background(), filter, fare, findRepOpts).Decode(&fare)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("location_wise_fare.updated", &fare)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(fare.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return fare, nil
}

// DeleteLocationWiseFareByID deletes location wise fare by id.
func DeleteLocationWiseFareByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	companiesCollection := db.Collection(LocationWiseFareCollection)
	res, err := companiesCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("location_wise_fare.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (fare *LocationWiseFare) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, fare); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (fare *LocationWiseFare) MarshalBinary() ([]byte, error) {
	return json.Marshal(fare)
}
