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

// StoreVehicleType represents a store vehicle type.
type StoreVehicleType struct {
	ID                        primitive.ObjectID       `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt                 time.Time                `json:"createdAt" bson:"createdAt"`
	DeletedAt                 *time.Time               `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt                 time.Time                `json:"updatedAt" bson:"updatedAt"`
	CreatedBy                 primitive.ObjectID       `json:"createdBy" bson:"createdBy"`
	Type                      string                   `json:"type" bson:"type"`
	Location                  StoreVehicleTypeLocation `json:"location" bson:"location"`
	ChargesForCompletedOrders int                      `json:"chargesForCompletedOrders" bson:"chargesForCompletedOrders"`
	ChargesForCancelledOrders int                      `json:"chargesForCancelledOrders" bson:"chargesForCancelledOrders"`
	DeliveryRadius            int                      `json:"deliveryRadius" bson:"deliveryRadius"`
	Order                     int                      `json:"order" bson:"order"`
	IsActive                  bool                     `json:"isActive" bson:"isActive"`
}

// CreateStoreVehicleType creates store vehicle type.
func CreateStoreVehicleType(storeVehicleType StoreVehicleType) (*StoreVehicleType, error) {
	storeVehicleType.CreatedAt = time.Now()
	storeVehicleType.UpdatedAt = time.Now()
	storeVehicleType.ID = primitive.NewObjectID()
	db := database.MongoDB
	installationCollection := db.Collection(StoreVehicleTypesCollection)
	ctx := context.Background()
	_, err := installationCollection.InsertOne(ctx, &storeVehicleType)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("store_vehicle_type.created", &storeVehicleType)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(storeVehicleType.ID.Hex(), storeVehicleType, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &storeVehicleType, nil
}

// GetStoreVehicleTypeByID gives a store vehicle type by id.
func GetStoreVehicleTypeByID(ID string) (*StoreVehicleType, error) {
	db := database.MongoDB
	storeVehicleType := &StoreVehicleType{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(storeVehicleType)
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
	err = db.Collection(StoreVehicleTypesCollection).FindOne(ctx, filter).Decode(&storeVehicleType)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, storeVehicleType, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return storeVehicleType, nil
}

// GetStoreVehicleTypes gives a list of store vehicle types.
func GetStoreVehicleTypes(filter bson.D, limit int, after *string, before *string, first *int, last *int) (storeVehicleTypes []*StoreVehicleType, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(StoreVehicleTypesCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(StoreVehicleTypesCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
ctx := context.Background()
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		storeVehicleType := &StoreVehicleType{}
		err = cur.Decode(&storeVehicleType)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		storeVehicleTypes = append(storeVehicleTypes, storeVehicleType)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return storeVehicleTypes, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateStoreVehicleType updates store vehicle type.
func UpdateStoreVehicleType(s *StoreVehicleType) (*StoreVehicleType, error) {
	storeVehicleType := s
	storeVehicleType.UpdatedAt = time.Now()
	filter := bson.D{{"_id", storeVehicleType.ID}}
	db := database.MongoDB
	storeVehicleTypesCollection := db.Collection(StoreVehicleTypesCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := storeVehicleTypesCollection.FindOneAndReplace(context.Background(), filter, storeVehicleType, findRepOpts).Decode(&storeVehicleType)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("store_vehicle_type.updated", &storeVehicleType)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(storeVehicleType.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return storeVehicleType, nil
}

// DeleteStoreVehicleTypeByID deletes store vehicle type by id.
func DeleteStoreVehicleTypeByID(ID string) (bool, error) {
	db := database.MongoDB
	id, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", id}}
	storeVehicleTypesCollection := db.Collection(StoreVehicleTypesCollection)
	res, err := storeVehicleTypesCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	log.Warningln(res)
	if err != nil && res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("store_vehicle_type.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (storeVehicleType *StoreVehicleType) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, storeVehicleType); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (storeVehicleType *StoreVehicleType) MarshalBinary() ([]byte, error) {
	return json.Marshal(storeVehicleType)
}
