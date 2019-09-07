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

// Notification represents a notification.
type Notification struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`
	IsActive  bool               `json:"isActive" bson:"isActive"`
}

// CreateNotification creates new notification.
func CreateNotification(notification Notification) (*Notification, error) {
	notification.CreatedAt = time.Now()
	notification.UpdatedAt = time.Now()
	notification.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(NotificationsCollection)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := collection.InsertOne(ctx, &notification)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("notification.created", &notification)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(notification.ID.Hex(), notification, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &notification, nil
}

// GetNotificationByID gives requested notification by id.
func GetNotificationByID(ID string) (*Notification, error) {
	db := database.MongoDB
	notification := &Notification{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(notification)
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
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	err = db.Collection(NotificationsCollection).FindOne(ctx, filter).Decode(&notification)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, notification, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return notification, nil
}

// GetNotifications gives a list of notifications.
func GetNotifications(filter bson.D, limit int, after *string, before *string, first *int, last *int) (notifications []*Notification, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(NotificationsCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(NotificationsCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		notification := &Notification{}
		err = cur.Decode(&notification)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		notifications = append(notifications, notification)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return notifications, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateNotification updates notifications.
func UpdateNotification(c *Notification) (*Notification, error) {
	notification := c
	notification.UpdatedAt = time.Now()
	filter := bson.D{{"_id", notification.ID}}
	db := database.MongoDB
	notificationsCollection := db.Collection(NotificationsCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := notificationsCollection.FindOneAndReplace(context.Background(), filter, notification, findRepOpts).Decode(&notification)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("notification.updated", &notification)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(notification.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return notification, nil
}

// DeleteNotificationByID deletes notifications by id.
func DeleteNotificationByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	notificationsCollection := db.Collection(NotificationsCollection)
	res, err := notificationsCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("notification.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (notification *Notification) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, notification); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (notification *Notification) MarshalBinary() ([]byte, error) {
	return json.Marshal(notification)
}

// PushNotification represents push notification.
type PushNotification struct {
	ID      primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Success bool               `json:"success" bson:"success"`
	Message string             `json:"message" bson:"message"`
}

// GetPushNotificationByID returns the requested push notification by id.
func GetPushNotificationByID(ID string) (*PushNotification, error) {
	db := database.MongoDB
	notification := &PushNotification{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(notification)
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
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	err = db.Collection(PushNotificationsCollection).FindOne(ctx, filter).Decode(&notification)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, notification, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return notification, nil
}

//UnmarshalBinary required for the redis cache to work
func (notification *PushNotification) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, notification); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (notification *PushNotification) MarshalBinary() ([]byte, error) {
	return json.Marshal(notification)
}
