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

// DeliveryVehicleType represents a delivery vehicle type.
type DeliveryVehicleType struct {
	ID                              primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt                       time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt                       *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt                       time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy                       primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	VehicleType                     string             `json:"vehicleType" bson:"vehicleType"`
	Location                        string             `json:"location" bson:"location"`
	DeliveryChargeForCompletedOrder float64            `json:"deliveryChargeForCompletedOrder" bson:"deliveryChargeForCompletedOrder"`
	DeliveryChargeForCancelledOrder float64            `json:"deliveryChargeForCancelledOrder" bson:"deliveryChargeForCancelledOrder"`
	DeliveryRadius                  float64            `json:"deliveryRadius" bson:"deliveryRadius"`
	Order                           int                `json:"order" bson:"order"`
	IsActive                        bool               `json:"isActive" bson:"isActive"`
}

// CreateDeliveryVehicleType creates new delivery vehicle types.
func CreateDeliveryVehicleType(deliveryVehicleType DeliveryVehicleType) (*DeliveryVehicleType, error) {
	deliveryVehicleType.CreatedAt = time.Now()
	deliveryVehicleType.UpdatedAt = time.Now()
	deliveryVehicleType.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(DeliveryVehicleTypeCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &deliveryVehicleType)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("delivery_vehicle_type.created", &deliveryVehicleType)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(deliveryVehicleType.ID.Hex(), deliveryVehicleType, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &deliveryVehicleType, nil
}

// GetDeliveryVehicleTypeByID gives the requested delivery vehicle type using id.
func GetDeliveryVehicleTypeByID(ID string) (*DeliveryVehicleType, error) {
	db := database.MongoDB
	deliveryVehicleType := &DeliveryVehicleType{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(deliveryVehicleType)
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
	err = db.Collection(DeliveryVehicleTypeCollection).FindOne(ctx, filter).Decode(&deliveryVehicleType)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, deliveryVehicleType, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return deliveryVehicleType, nil
}

// GetDeliveryVehicleTypes gives an array of delivery vehicle types.
func GetDeliveryVehicleTypes(filter bson.D, limit int, after *string, before *string, first *int, last *int) (deliveryVehicleTypes []*DeliveryVehicleType, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(DeliveryVehicleTypeCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(DeliveryVehicleTypeCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		deliveryVehicleType := &DeliveryVehicleType{}
		err = cur.Decode(&deliveryVehicleType)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		deliveryVehicleTypes = append(deliveryVehicleTypes, deliveryVehicleType)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return deliveryVehicleTypes, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateDeliveryVehicleType updates the delivery vehicle type.
func UpdateDeliveryVehicleType(c *DeliveryVehicleType) (*DeliveryVehicleType, error) {
	deliveryVehicleType := c
	deliveryVehicleType.UpdatedAt = time.Now()
	filter := bson.D{{"_id", deliveryVehicleType.ID}}
	db := database.MongoDB
	deliveryVehicleTypesCollection := db.Collection(DeliveryVehicleTypeCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := deliveryVehicleTypesCollection.FindOneAndReplace(context.Background(), filter, deliveryVehicleType, findRepOpts).Decode(&deliveryVehicleType)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("delivery_vehicle_type.updated", &deliveryVehicleType)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(deliveryVehicleType.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return deliveryVehicleType, nil
}

// DeleteDeliveryVehicleTypeByID deletes the delivery vehicle type by id.
func DeleteDeliveryVehicleTypeByID(ID string) (bool, error) {
	db := database.MongoDB
	id, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", id}}
	deliveryVehicleTypesCollection := db.Collection(DeliveryVehicleTypeCollection)
	res, err := deliveryVehicleTypesCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("delivery_vehicle_type.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (deliveryVehicleType *DeliveryVehicleType) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, deliveryVehicleType); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (deliveryVehicleType *DeliveryVehicleType) MarshalBinary() ([]byte, error) {
	return json.Marshal(deliveryVehicleType)
}
