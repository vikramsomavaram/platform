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

// DeliveryChargesUtility represents delivery charge utility.
type DeliveryChargesUtility struct {
	ID                              primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt                       time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt                       *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt                       time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy                       primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Location                        string             `json:"location" bson:"location"`
	OrderPrice                      int                `json:"orderPrice" bson:"orderPrice"`
	OrderDeliveryChargesAboveAmout  int                `json:"orderDeliveryChargesAboveAmout" bson:"orderDeliveryChargesAboveAmout"`
	OrderDeliveryChargesBelowAmount int                `json:"orderDeliveryChargesBelowAmount" bson:"orderDeliveryChargesBelowAmount"`
	FreeOrderDeliveryCharges        int                `json:"freeOrderDeliveryCharges" bson:"freeOrderDeliveryCharges"`
	FreeDeliveryRadius              int                `json:"freeDeliveryRadius" bson:"freeDeliveryRadius"`
	IsActive                        bool               `json:"isActive" bson:"isActive"`
}

// CreateDeliveryChargesUtility creates new delivery charge utility.
func CreateDeliveryChargesUtility(deliveryChargesUtility DeliveryChargesUtility) (*DeliveryChargesUtility, error) {
	deliveryChargesUtility.CreatedAt = time.Now()
	deliveryChargesUtility.UpdatedAt = time.Now()
	deliveryChargesUtility.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(DeliveryChargeUtilitiesCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &deliveryChargesUtility)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("delivery_charge_utility.created", &deliveryChargesUtility)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(deliveryChargesUtility.ID.Hex(), deliveryChargesUtility, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &deliveryChargesUtility, nil
}

// GetDeliveryChargesUtilityByID returns requested delivery charge utility by id.
func GetDeliveryChargesUtilityByID(ID string) (*DeliveryChargesUtility, error) {
	db := database.MongoDB
	deliveryChargesUtility := &DeliveryChargesUtility{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(deliveryChargesUtility)
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
	err = db.Collection(DeliveryChargeUtilitiesCollection).FindOne(ctx, filter).Decode(&deliveryChargesUtility)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, deliveryChargesUtility, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return deliveryChargesUtility, nil
}

// GetDeliveryChargesUtilities returns a list of delivery charge utilities.
func GetDeliveryChargesUtilities(filter bson.D, limit int, after *string, before *string, first *int, last *int) (deliveryChargesUtilities []*DeliveryChargesUtility, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(DeliveryChargeUtilitiesCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(DeliveryChargeUtilitiesCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		deliveryChargesUtility := &DeliveryChargesUtility{}
		err = cur.Decode(&deliveryChargesUtility)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		deliveryChargesUtilities = append(deliveryChargesUtilities, deliveryChargesUtility)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return deliveryChargesUtilities, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateDeliveryChargesUtility updates delivery charge utility.
func UpdateDeliveryChargesUtility(c *DeliveryChargesUtility) (*DeliveryChargesUtility, error) {
	deliveryChargesUtility := c
	deliveryChargesUtility.UpdatedAt = time.Now()
	filter := bson.D{{"_id", deliveryChargesUtility.ID}}
	db := database.MongoDB
	deliveryChargeUtilitiesCollection := db.Collection(DeliveryChargeUtilitiesCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := deliveryChargeUtilitiesCollection.FindOneAndReplace(context.Background(), filter, deliveryChargesUtility, findRepOpts).Decode(&deliveryChargesUtility)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("delivery_charge_utility.updated", &deliveryChargesUtility)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(deliveryChargesUtility.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return deliveryChargesUtility, nil
}

// DeleteDeliveryChargesUtilityByID deletes delivery charge utility by id.
func DeleteDeliveryChargesUtilityByID(ID string) (bool, error) {
	db := database.MongoDB
	id, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", id}}
	deliveryChargeUtilitiesCollection := db.Collection(DeliveryChargeUtilitiesCollection)
	res, err := deliveryChargeUtilitiesCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("delivery_charge_utility.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (deliveryChargesUtility *DeliveryChargesUtility) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, deliveryChargesUtility); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (deliveryChargesUtility *DeliveryChargesUtility) MarshalBinary() ([]byte, error) {
	return json.Marshal(deliveryChargesUtility)
}
