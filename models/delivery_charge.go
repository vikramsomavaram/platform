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

// DeliveryCharge represents a delivery charge.
type DeliveryCharge struct {
	ID                              primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt                       time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt                       *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt                       time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy                       primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	LocationName                    string             `json:"locationName" bson:"locationName"`
	OrderPrice                      int                `json:"orderPrice" bson:"orderPrice"`
	OrderDeliveryChargesAboveAmount int                `json:"orderDeliveryChargesAboveAmount" bson:"orderDeliveryChargesAboveAmount"`
	OrderDeliveryChargesBelowAmount int                `json:"orderDeliveryChargesBelowAmount" bson:"orderDeliveryChargesBelowAmount"`
	FreeOrderDeliveryCharges        int                `json:"freeOrderDeliveryCharges" bson:"freeOrderDeliveryCharges"`
	FreeDeliveryRadius              int                `json:"freeDeliveryRadius" bson:"freeDeliveryRadius"`
	OrderTotal                      int                `json:"orderTotal" bson:"orderTotal"`
	IsActive                        bool               `json:"isActive" bson:"isActive"`
}

// CreateDeliveryCharge creates new delivery charge.
func CreateDeliveryCharge(deliveryCharge DeliveryCharge) (*DeliveryCharge, error) {
	deliveryCharge.CreatedAt = time.Now()
	deliveryCharge.UpdatedAt = time.Now()
	deliveryCharge.ID = primitive.NewObjectID()
	db := database.MongoDB
	Collection := db.Collection(DeliveryChargesCollection)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := Collection.InsertOne(ctx, &deliveryCharge)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("delivery_charge.created", &deliveryCharge)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(deliveryCharge.ID.Hex(), deliveryCharge, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &deliveryCharge, nil
}

// GetDeliveryChargeByID gives the requested delivery charge using id.
func GetDeliveryChargeByID(ID string) (*DeliveryCharge, error) {
	db := database.MongoDB
	deliveryCharge := &DeliveryCharge{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(deliveryCharge)
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
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	err = db.Collection(DeliveryChargesCollection).FindOne(ctx, filter).Decode(&deliveryCharge)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, deliveryCharge, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return deliveryCharge, nil
}

// GetDeliveryCharges gives an array of delivery charges.
func GetDeliveryCharges(filter bson.D, limit int, after *string, before *string, first *int, last *int) (deliveryCharges []*DeliveryCharge, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(DeliveryChargesCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(DeliveryChargesCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		deliveryCharge := &DeliveryCharge{}
		err = cur.Decode(&deliveryCharge)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		deliveryCharges = append(deliveryCharges, deliveryCharge)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return deliveryCharges, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateDeliveryCharge updates the delivery charges.
func UpdateDeliveryCharge(d *DeliveryCharge) (*DeliveryCharge, error) {
	deliveryCharge := d
	deliveryCharge.UpdatedAt = time.Now()
	filter := bson.D{{"_id", deliveryCharge.ID}}
	db := database.MongoDB
	deliveryChargesCollection := db.Collection(DeliveryChargesCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := deliveryChargesCollection.FindOneAndReplace(context.Background(), filter, deliveryCharge, findRepOpts).Decode(&deliveryCharge)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("delivery_charge.updated", &deliveryCharge)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(deliveryCharge.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return deliveryCharge, nil
}

// DeleteDeliveryChargeByID deletes the delivery charge using id.
func DeleteDeliveryChargeByID(ID string) (bool, error) {
	db := database.MongoDB
	id, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", id}}
	deliveryChargesCollection := db.Collection(DeliveryChargesCollection)
	res, err := deliveryChargesCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil && res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("delivery_charge.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (deliveryCharge *DeliveryCharge) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, deliveryCharge); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (deliveryCharge *DeliveryCharge) MarshalBinary() ([]byte, error) {
	return json.Marshal(deliveryCharge)
}
