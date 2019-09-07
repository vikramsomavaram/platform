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

// AirportSurcharge represents a of airport surcharge.
type AirportSurcharge struct {
	ID               primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt        time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt        *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt        time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy        primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	AirportLocation  string             `json:"airportLocation" bson:"airportLocation"`
	PickUpSurcharge  string             `json:"pickUpSurcharge" bson:"pickUpSurcharge"`
	DropOffSurcharge string             `json:"dropOffSurcharge" bson:"dropOffSurcharge"`
	VehicleType      string             `json:"vehicleType" bson:"vehicleType"`
	IsActive         bool               `json:"isActive" bson:"isActive"`
}

// CreateAirportSurcharge creates airport surcharge.
func CreateAirportSurcharge(airportSurcharge AirportSurcharge) (*AirportSurcharge, error) {
	airportSurcharge.CreatedAt = time.Now()
	airportSurcharge.UpdatedAt = time.Now()
	airportSurcharge.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(AirportSurchargeCollection)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := collection.InsertOne(ctx, &airportSurcharge)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("airport_surcharge.created", &airportSurcharge)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(airportSurcharge.ID.Hex(), airportSurcharge, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &airportSurcharge, nil
}

// GetAirportSurchargeByID gives airport surcharge by id.
func GetAirportSurchargeByID(ID string) (*AirportSurcharge, error) {
	db := database.MongoDB
	airportSurcharge := &AirportSurcharge{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(airportSurcharge)
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
	err = db.Collection(AirportSurchargeCollection).FindOne(context.Background(), filter).Decode(&airportSurcharge)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, airportSurcharge, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return airportSurcharge, nil
}

// GetAirportSurcharges gives a list of airport surcharges.
func GetAirportSurcharges(filter bson.D, limit int, after *string, before *string, first *int, last *int) (airportSurcharges []*AirportSurcharge, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(AirportSurchargeCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(AirportSurchargeCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		airportSurcharge := &AirportSurcharge{}
		err = cur.Decode(&airportSurcharge)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		airportSurcharges = append(airportSurcharges, airportSurcharge)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return airportSurcharges, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateAirportSurcharge updates airport surcharge.
func UpdateAirportSurcharge(c *AirportSurcharge) (*AirportSurcharge, error) {
	airportSurcharge := c
	airportSurcharge.UpdatedAt = time.Now()
	filter := bson.D{{"_id", airportSurcharge.ID}}
	db := database.MongoDB
	airportSurchargeCollection := db.Collection(AirportSurchargeCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := airportSurchargeCollection.FindOneAndReplace(context.Background(), filter, airportSurcharge, findRepOpts).Decode(&airportSurcharge)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("airport_surcharge.updated", &airportSurcharge)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(airportSurcharge.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return airportSurcharge, nil
}

// DeleteAirportSurchargeByID deletes airport surcharge.
func DeleteAirportSurchargeByID(ID string) (bool, error) {
	db := database.MongoDB
	id, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", id}}
	airportSurchargeCollection := db.Collection(AirportSurchargeCollection)
	res, err := airportSurchargeCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("airport_surcharge.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (airportSurcharge *AirportSurcharge) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, airportSurcharge); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (airportSurcharge *AirportSurcharge) MarshalBinary() ([]byte, error) {
	return json.Marshal(airportSurcharge)
}
