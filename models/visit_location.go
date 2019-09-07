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

// VisitLocation represents a visit location.
type VisitLocation struct {
	ID               primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt        time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt        *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt        time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy        primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	DestinationTitle string             `json:"destinationTitle" bson:"destinationTitle"`
	Destination      string             `json:"destination" bson:"destination"`
	IsActive         bool               `json:"isActive" bson:"isActive"`
}

// CreateVisitLocation creates visit location.
func CreateVisitLocation(visitLocation VisitLocation) (*VisitLocation, error) {
	visitLocation.CreatedAt = time.Now()
	visitLocation.UpdatedAt = time.Now()
	visitLocation.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(VisitLocationCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &visitLocation)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("visit_location.created", &visitLocation)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(visitLocation.ID.Hex(), visitLocation, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &visitLocation, nil
}

// GetVisitLocationByID gives visit location by id.
func GetVisitLocationByID(ID string) (*VisitLocation, error) {
	db := database.MongoDB
	visitLocation := &VisitLocation{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(visitLocation)
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
	err = db.Collection(VisitLocationCollection).FindOne(context.Background(), filter).Decode(&visitLocation)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, visitLocation, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return visitLocation, nil
}

// GetVisitLocations gives a list of visit locations.
func GetVisitLocations(filter bson.D, limit int, after *string, before *string, first *int, last *int) (visitLocations []*VisitLocation, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(VisitLocationCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(VisitLocationCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
ctx := context.Background()
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		visitLocation := &VisitLocation{}
		err = cur.Decode(&visitLocation)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		visitLocations = append(visitLocations, visitLocation)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return visitLocations, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateVisitLocation updates visit location.
func UpdateVisitLocation(c *VisitLocation) (*VisitLocation, error) {
	visitLocation := c
	visitLocation.UpdatedAt = time.Now()
	filter := bson.D{{"_id", visitLocation.ID}}
	db := database.MongoDB
	visitLocationCollection := db.Collection(VisitLocationCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := visitLocationCollection.FindOneAndReplace(context.Background(), filter, visitLocation, findRepOpts).Decode(&visitLocation)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("visit_location.updated", &visitLocation)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(visitLocation.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return visitLocation, nil
}

// DeleteVisitLocationByID deletes visit location by id.
func DeleteVisitLocationByID(ID string) (bool, error) {
	db := database.MongoDB
	id, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", id}}
	visitLocationCollection := db.Collection(VisitLocationCollection)
	res, err := visitLocationCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("visit_location.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (visitLocation *VisitLocation) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, visitLocation); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (visitLocation *VisitLocation) MarshalBinary() ([]byte, error) {
	return json.Marshal(visitLocation)
}
