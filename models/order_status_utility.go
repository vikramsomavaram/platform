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

// OrderStatusUtility represents a order status utility.
type OrderStatusUtility struct {
	ID                primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt         time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt         *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt         time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy         primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	StatusTitle       string             `json:"statusTitle" bson:"statusTitle"`
	StatusDescription string             `json:"statusDescription" bson:"statusDescription"`
	IsActive          bool               `json:"isActive" bson:"isActive"`
}

// CreateOrderStatusUtility creates new order status utility.
func CreateOrderStatusUtility(orderStatusUtility OrderStatusUtility) (*OrderStatusUtility, error) {
	orderStatusUtility.CreatedAt = time.Now()
	orderStatusUtility.UpdatedAt = time.Now()
	orderStatusUtility.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(OrderStatusUtilityCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &orderStatusUtility)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("order_status_utility.created", &orderStatusUtility)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(orderStatusUtility.ID.Hex(), orderStatusUtility, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &orderStatusUtility, nil
}

// GetOrderStatusUtilityByID gives requested order status utility by id.
func GetOrderStatusUtilityByID(ID string) (*OrderStatusUtility, error) {
	db := database.MongoDB
	orderStatusUtility := &OrderStatusUtility{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(orderStatusUtility)
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
	err = db.Collection(OrderStatusUtilityCollection).FindOne(ctx, filter).Decode(&orderStatusUtility)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, orderStatusUtility, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return orderStatusUtility, nil
}

// GetOrderstatusUtilities gives a list of order status utilities.
func GetOrderstatusUtilities(filter bson.D, limit int, after *string, before *string, first *int, last *int) (orderSatusUtilities []*OrderStatusUtility, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(OrderStatusUtilityCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(OrderStatusUtilityCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
ctx := context.Background()
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		orderSatusUtility := &OrderStatusUtility{}
		err = cur.Decode(&orderSatusUtility)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		orderSatusUtilities = append(orderSatusUtilities, orderSatusUtility)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return orderSatusUtilities, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateOrderStatusUtility updates order status utility.
func UpdateOrderStatusUtility(c *OrderStatusUtility) (*OrderStatusUtility, error) {
	orderStatusUtility := c
	orderStatusUtility.UpdatedAt = time.Now()
	filter := bson.D{{"_id", orderStatusUtility.ID}}
	db := database.MongoDB
	orderStatusUtilityCollection := db.Collection(OrderStatusUtilityCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := orderStatusUtilityCollection.FindOneAndReplace(context.Background(), filter, orderStatusUtility, findRepOpts).Decode(&orderStatusUtility)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("order_status_utility.updated", &orderStatusUtility)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(orderStatusUtility.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return orderStatusUtility, nil
}

// DeleteOrderStatusUtilityByID deletes order status utility by id.
func DeleteOrderStatusUtilityByID(ID string) (bool, error) {
	db := database.MongoDB
	id, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", id}}
	orderStatusUtilityCollection := db.Collection(OrderStatusUtilityCollection)
	res, err := orderStatusUtilityCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("order_status_utility.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (orderStatusUtility *OrderStatusUtility) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, orderStatusUtility); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (orderStatusUtility *OrderStatusUtility) MarshalBinary() ([]byte, error) {
	return json.Marshal(orderStatusUtility)
}
