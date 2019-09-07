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

// Webhook represents a developer webhook.
type Webhook struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt   time.Time          `json:"-" bson:"deletedAt,omitempty"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy   primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	AppID       string             `json:"appId" bson:"appId"`
	URL         string             `json:"url" bson:"url"`
	EventTopics []string           `json:"events" bson:"events"`
	Secret      string             `json:"secret" bson:"secret"`
	IsActive    bool               `json:"isActive" bson:"isActive"`
}

//WebhookStatistics represents webhook statistics.
type WebhookStatistics struct {
	Stats string `json:"stats" bson:"stats"`
}

//UnmarshalBinary required for the redis cache to work
func (wh *Webhook) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, wh); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (wh *Webhook) MarshalBinary() ([]byte, error) {
	return json.Marshal(wh)
}

// CreateWebhook creates webhook.
func CreateWebhook(webhook Webhook) (*Webhook, error) {
	webhook.CreatedAt = time.Now()
	webhook.UpdatedAt = time.Now()
	webhook.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(WebhooksCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &webhook)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("webhook.created", &webhook)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(webhook.ID.Hex(), webhook, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &webhook, nil
}

// GetWebhookByID gives webhook by id.
func GetWebhookByID(ID string) (*Webhook, error) {
	db := database.MongoDB
	webhook := &Webhook{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(webhook)
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
	err = db.Collection(WebhooksCollection).FindOne(ctx, filter).Decode(&webhook)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, webhook, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return webhook, nil
}

// GetWebhooks gives a list of webhooks.
func GetWebhooks(filter bson.D, limit int, after *string, before *string, first *int, last *int) (webhooks []*Webhook, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(WebhooksCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(WebhooksCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		webhook := &Webhook{}
		err = cur.Decode(&webhook)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		webhooks = append(webhooks, webhook)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return webhooks, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateWebhook updates webhook.
func UpdateWebhook(c *Webhook) (*Webhook, error) {
	webhook := c
	webhook.UpdatedAt = time.Now()
	filter := bson.D{{"_id", webhook.ID}}
	db := database.MongoDB
	webhooksCollection := db.Collection(WebhooksCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := webhooksCollection.FindOneAndReplace(context.Background(), filter, webhook, findRepOpts).Decode(&webhook)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("webhook.updated", &webhook)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(webhook.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return webhook, nil
}

// DeleteWebhookByID deletes webhook by id.
func DeleteWebhookByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, _ := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	webhooksCollection := db.Collection(WebhooksCollection)
	res, err := webhooksCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("webhook.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

// WebhookLog represents a webhooklog.
type WebhookLog struct {
	ID                   primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	WebhookID            string             `json:"webhookId" bson:"webhookId"`
	CreatedAt            time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt            time.Time          `json:"-" bson:"deletedAt,omitempty"`
	UpdatedAt            time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy            string             `json:"createdBy" bson:"createdBy"`
	EventType            string             `json:"eventType" bson:"eventType"`
	Payload              string             `json:"payload" bson:"payload"`
	ClientResponseStatus string             `json:"clientResponseStatus" bson:"clientResponseStatus"`
	ClientResponseCode   int                `json:"clientResponseCode" bson:"clientResponseCode"`
}

// GetWebhookLogByID gives a webhook log by id.
func GetWebhookLogByID(ID string) (*WebhookLog, error) {
	db := database.MongoDB
	webhookLog := &WebhookLog{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(webhookLog)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	oID, _ := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}, {"deletedAt", bson.M{"$exists": false}}}
	err = db.Collection(WebhookLogsCollection).FindOne(context.Background(), filter).Decode(&webhookLog)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, webhookLog, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return webhookLog, nil
}

// GetWebhookLogs gives a list of webhook logs.
func GetWebhookLogs(filter bson.D, limit int, after *string, before *string, first *int, last *int) (webhookLogs []*WebhookLog, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(WebhookLogsCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(WebhookLogsCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		webhookLog := &WebhookLog{}
		err = cur.Decode(&webhookLog)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		webhookLogs = append(webhookLogs, webhookLog)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return webhookLogs, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

//UnmarshalBinary required for the redis cache to work
func (webhookLog *WebhookLog) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, webhookLog); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (webhookLog *WebhookLog) MarshalBinary() ([]byte, error) {
	return json.Marshal(webhookLog)
}
