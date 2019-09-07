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
	"time"
)

// NewsletterSubscriber represents a news letter subscriber.
type NewsletterSubscriber struct {
	ID        primitive.ObjectID         `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt time.Time                  `json:"createdAt" bson:"createdAt"`
	DeletedAt *time.Time                 `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt time.Time                  `json:"updatedAt" bson:"updatedAt"`
	CreatedBy string                     `json:"createdBy" bson:"createdBy"`
	Name      string                     `json:"name" bson:"name"`
	Email     string                     `json:"email" bson:"email"`
	Status    NewsletterSubscriberStatus `json:"status" bson:"status"`
	Date      time.Time                  `json:"date" bson:"date"`
	IPAddress string                     `json:"ipAddress" bson:"ipAddress"`
	IsActive  bool                       `json:"isActive" bson:"isActive"`
}

// CreateNewsletterSubscriber creates new news letter subscriber.
func CreateNewsletterSubscriber(newsletterSubscriber NewsletterSubscriber) (*NewsletterSubscriber, error) {
	newsletterSubscriber.CreatedAt = time.Now()
	newsletterSubscriber.UpdatedAt = time.Now()
	newsletterSubscriber.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(NewsletterSubscribersCollection)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := collection.InsertOne(ctx, &newsletterSubscriber)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("newsletter_subscriber.created", &newsletterSubscriber)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(newsletterSubscriber.ID.Hex(), newsletterSubscriber, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &newsletterSubscriber, nil
}

// GetNewsletterSubscriberByID gives requested news letter subscriber by id.
func GetNewsletterSubscriberByID(ID string) (*NewsletterSubscriber, error) {
	db := database.MongoDB
	newsletterSubscriber := &NewsletterSubscriber{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(newsletterSubscriber)
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
	err = db.Collection(NewsletterSubscribersCollection).FindOne(ctx, filter).Decode(&newsletterSubscriber)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, newsletterSubscriber, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return newsletterSubscriber, nil
}

// GetNewsletterSubscribers gives a list of news letter subscribers.
func GetNewsletterSubscribers(filter bson.D, limit int, after *string, before *string, first *int, last *int) (newsletterSubscribers []*NewsletterSubscriber, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(NewsletterSubscribersCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(NewsletterSubscribersCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		newsletterSubscriber := &NewsletterSubscriber{}
		err = cur.Decode(&newsletterSubscriber)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		newsletterSubscribers = append(newsletterSubscribers, newsletterSubscriber)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return newsletterSubscribers, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

//UnmarshalBinary required for the redis cache to work
func (newsletterSubscriber *NewsletterSubscriber) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, newsletterSubscriber); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (newsletterSubscriber *NewsletterSubscriber) MarshalBinary() ([]byte, error) {
	return json.Marshal(newsletterSubscriber)
}
