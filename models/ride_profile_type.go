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

// RideProfileType represents a ride profile type.
type RideProfileType struct {
	ID               primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt        time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt        *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt        time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy        primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	ProfileShortName string             `json:"profileShortName" bson:"profileShortName"`
	OrganizationType string             `json:"organizationType" bson:"organizationType"`
	ProfileTitle     string             `json:"profileTitle" bson:"profileTitle"`
	TitleDescription string             `json:"titleDescription" bson:"titleDescription"`
	ScreenHeading    string             `json:"screenHeading" bson:"screenHeading"`
	ScreenTitle      string             `json:"screenTitle" bson:"screenTitle"`
	ButtonText       string             `json:"buttonText" bson:"buttonText"`
	ProfileIcon      string             `json:"profileIcon" bson:"profileIcon"`
	WelcomePicture   string             `json:"welcomePicture" bson:"welcomePicture"`
	IsActive         bool               `json:"isActive" bson:"isActive"`
}

// CreateRideProfileType creates new ride profile type.
func CreateRideProfileType(rideProfileType RideProfileType) (*RideProfileType, error) {
	rideProfileType.CreatedAt = time.Now()
	rideProfileType.UpdatedAt = time.Now()
	rideProfileType.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(RideProfileTypeCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &rideProfileType)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("ride_profile_type.created", &rideProfileType)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(rideProfileType.ID.Hex(), rideProfileType, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &rideProfileType, nil
}

// GetRideProfileTypeByID gives ride profile type by id.
func GetRideProfileTypeByID(ID string) (*RideProfileType, error) {
	db := database.MongoDB
	rideProfileType := &RideProfileType{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(rideProfileType)
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
	err = db.Collection(RideProfileTypeCollection).FindOne(ctx, filter).Decode(&rideProfileType)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, rideProfileType, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return rideProfileType, nil
}

// GetRideProfileTypes gives a list of ride profile type.
func GetRideProfileTypes(filter bson.D, limit int, after *string, before *string, first *int, last *int) (rideProfileTypes []*RideProfileType, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(RideProfileTypeCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(RideProfileTypeCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		rideProfileType := &RideProfileType{}
		err = cur.Decode(&rideProfileType)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		rideProfileTypes = append(rideProfileTypes, rideProfileType)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return rideProfileTypes, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateRideProfileType updates ride profile types.
func UpdateRideProfileType(c *RideProfileType) (*RideProfileType, error) {
	rideProfileType := c
	rideProfileType.UpdatedAt = time.Now()
	filter := bson.D{{"_id", rideProfileType.ID}}
	db := database.MongoDB
	rideProfileTypeCollection := db.Collection(RideProfileTypeCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := rideProfileTypeCollection.FindOneAndReplace(context.Background(), filter, rideProfileType, findRepOpts).Decode(&rideProfileType)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("ride_profile_type.updated", &rideProfileType)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(rideProfileType.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return rideProfileType, nil
}

// DeleteRideProfileTypeByID deletes ride profile type by id.
func DeleteRideProfileTypeByID(ID string) (bool, error) {
	db := database.MongoDB
	id, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", id}}
	rideProfileTypeCollection := db.Collection(RideProfileTypeCollection)
	res, err := rideProfileTypeCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("ride_profile_type.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (rideProfileType *RideProfileType) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, rideProfileType); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (rideProfileType *RideProfileType) MarshalBinary() ([]byte, error) {
	return json.Marshal(rideProfileType)
}
