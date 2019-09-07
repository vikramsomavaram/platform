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

// BusinessTripReason represents a business trip reason.
type BusinessTripReason struct {
	ID               primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt        time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt        *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt        time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy        primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	TripReason       string             `json:"tripReason" bson:"tripReason"`
	ProfileShortName string             `json:"profileShortName" bson:"profileShortName"`
	OrganizationType string             `json:"organizationType" bson:"organizationType"`
	ProfileTitle     string             `json:"profileTitle" bson:"profileTitle"`
	TitleDescription string             `json:"titleDescription" bson:"titleDescription"`
	Reason           string             `json:"reason" bson:"reason"`
	IsActive         bool               `json:"isActive" bson:"isActive"`
}

// CreateBusinessTripReason creates new business trip reason.
func CreateBusinessTripReason(businessTripReason BusinessTripReason) (*BusinessTripReason, error) {
	businessTripReason.CreatedAt = time.Now()
	businessTripReason.UpdatedAt = time.Now()
	businessTripReason.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(BusinessTripReasonCollection)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := collection.InsertOne(ctx, &businessTripReason)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(businessTripReason.ID.Hex(), businessTripReason, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("business_trip_reason.created", &businessTripReason)
	return &businessTripReason, nil
}

// GetBusinessTripReasonByID gives the requested business trip reason by id.
func GetBusinessTripReasonByID(ID string) *BusinessTripReason {
	db := database.MongoDB
	businessTripReason := &BusinessTripReason{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(businessTripReason)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	id, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		log.Errorln(err)
		return businessTripReason
	}
	filter := bson.D{{"_id", id}, {"deletedAt", bson.M{"$exists": false}}}
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	err = db.Collection(BusinessTripReasonCollection).FindOne(ctx, filter).Decode(&businessTripReason)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return businessTripReason
		}
		log.Errorln(err)
		return businessTripReason
	}
	//set cache item
	err = cacheClient.Set(ID, businessTripReason, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return businessTripReason
}

// GetBusinessTripReasons gives the array of business trip reasons.
func GetBusinessTripReasons(filter bson.D, limit int, after *string, before *string, first *int, last *int) (businessTripReasons []*BusinessTripReason, totalCount int64, hasPrevious, hasNext bool, err error) {
	db := database.MongoDB
	tcint, filter, err := calcTotalCountWithQueryFilters(BusinessTripReasonCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(BusinessTripReasonCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		businessTripReason := &BusinessTripReason{}
		err = cur.Decode(&businessTripReason)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		businessTripReasons = append(businessTripReasons, businessTripReason)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return businessTripReasons, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateBusinessTripReason updates the business trip reason.
func UpdateBusinessTripReason(c *BusinessTripReason) (*BusinessTripReason, error) {
	businessTripReason := c
	businessTripReason.UpdatedAt = time.Now()
	filter := bson.D{{"_id", businessTripReason.ID}}
	db := database.MongoDB
	businessTripReasonsCollection := db.Collection(BusinessTripReasonCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := businessTripReasonsCollection.FindOneAndReplace(context.Background(), filter, businessTripReason, findRepOpts).Decode(&businessTripReason)
	if err != nil {
		log.Error(err)
	}
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(businessTripReason.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("business_trip_reason.updated", &businessTripReason)
	return businessTripReason, nil
}

// DeleteBusinessTripReasonByID deletes the business trip reason by id.
func DeleteBusinessTripReasonByID(ID string) (bool, error) {
	db := database.MongoDB
	id, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", id}}
	businessTripReasonsCollection := db.Collection(BusinessTripReasonCollection)
	res, err := businessTripReasonsCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("business_trip_reason.deleted", &res)
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (businessTripReason *BusinessTripReason) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, businessTripReason); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (businessTripReason *BusinessTripReason) MarshalBinary() ([]byte, error) {
	return json.Marshal(businessTripReason)
}
