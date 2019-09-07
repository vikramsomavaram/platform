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

// DeclineAlert represents a decline alert.
type DeclineAlert struct {
	ID                         primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt                  time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt                  *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt                  time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy                  string             `json:"createdBy" bson:"createdBy"`
	ProviderName               string             `json:"providerName" bson:"providerName"`
	Email                      string             `json:"email" bson:"email"`
	TotalCancelledTrips        int                `json:"totalCancelledTrips" bson:"totalCancelledTrips"`
	TotalDeclinedTrips         int                `json:"totalDeclinedTrips" bson:"totalDeclinedTrips"`
	TotalCancelledTripsTillNow int                `json:"totalCancelledTripsTillNow" bson:"totalCancelledTripsTillNow"`
	TotalDeclinedTripsTillNow  int                `json:"totalDeclinedTripsTillNow" bson:"totalDeclinedTripsTillNow"`
	BlockProvider              bool               `json:"blockProvider" bson:"blockProvider"`
	BlockDate                  time.Time          `json:"blockDate" bson:"blockDate"`
	UserName                   string             `json:"userName" bson:"userName" `
	BlockUser                  bool               `json:"blockUser" bson:"blockUser"`
	IsActive                   bool               `json:"isActive" bson:"isActive"`
}

//DeclineAlertForProvider represents a decline alert for provider.
type DeclineAlertForProvider struct {
	ID                         primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt                  time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt                  *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt                  time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy                  string             `json:"createdBy" bson:"createdBy"`
	ProviderName               primitive.ObjectID `json:"providerName" bson:"providerName"`
	Email                      string             `json:"email" bson:"email"`
	TotalCancelledTrips        string             `json:"totalCancelledTrips" bson:"totalCancelledTrips"`
	TotalDeclinedTrips         string             `json:"totalDeclinedTrips" bson:"totalDeclinedTrips"`
	TotalCancelledTripsTillNow string             `json:"totalCancelledTripsTillNow" bson:"totalCancelledTripsTillNow"`
	TotalDeclinedTripsTillNow  string             `json:"totalDeclinedTripsTillNow" bson:"totalDeclinedTripsTillNow"`
	BlockProvider              bool               `json:"blockProvider" bson:"blockProvider"`
	BlockDate                  string             `json:"blockDate" bson:"blockDate"`
}

// DeclineAlertForUser represents a decline alert for user.
type DeclineAlertForUser struct {
	ID                         primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt                  time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt                  *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt                  time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy                  string             `json:"createdBy" bson:"createdBy"`
	UserName                   string             `json:"userName" bson:"userName"`
	Email                      string             `json:"email" bson:"email"`
	TotalCancelledTrips        string             `json:"totalCancelledTrips" bson:"totalCancelledTrips"`
	TotalCancelledTripsTillNow string             `json:"totalCancelledTripsTillNow" bson:"totalCancelledTripsTillNow"`
	BlockUser                  bool               `json:"blockUser" bson:"blockUser"`
	BlockDate                  string             `json:"blockDate" bson:"blockDate"`
}

// CreateDeclineAlert creates decline alert.
func CreateDeclineAlert(declineAlert DeclineAlert) (*DeclineAlert, error) {
	declineAlert.CreatedAt = time.Now()
	declineAlert.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(AlertsCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &declineAlert)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("decline_alert.created", &declineAlert)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(declineAlert.ID.Hex(), declineAlert, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &declineAlert, nil
}

// GetDeclineAlertForProviderByID gives the requested alert by id.
func GetDeclineAlertForProviderByID(ID string) *DeclineAlertForProvider {
	db := database.MongoDB
	declineAlertForProvider := &DeclineAlertForProvider{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(declineAlertForProvider)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}

	oID, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		log.Errorln(err)
		return declineAlertForProvider
	}
	filter := bson.D{{"_id", oID}, {"deletedAt", bson.M{"$exists": false}}}
	ctx := context.Background()
	err = db.Collection(AlertsCollection).FindOne(ctx, filter).Decode(&declineAlertForProvider)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return declineAlertForProvider
		}
		log.Errorln(err)
		return declineAlertForProvider
	}
	//set cache item
	err = cacheClient.Set(ID, declineAlertForProvider, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return declineAlertForProvider
}

// GetDeclineAlertForUserByID gives the requested alert by id.
func GetDeclineAlertForUserByID(ID string) *DeclineAlertForUser {
	db := database.MongoDB
	declineAlertForUser := &DeclineAlertForUser{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(declineAlertForUser)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	oID, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		log.Errorln(err)
		return declineAlertForUser
	}
	filter := bson.D{{"_id", oID}, {"deletedAt", bson.M{"$exists": false}}}
	ctx := context.Background()
	err = db.Collection(AlertsCollection).FindOne(ctx, filter).Decode(&declineAlertForUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return declineAlertForUser
		}
		log.Errorln(err)
		return declineAlertForUser
	}
	//set cache item
	err = cacheClient.Set(ID, declineAlertForUser, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return declineAlertForUser
}

// GetDeclineAlertsForUsers returns a list of decline alerts for users.
func GetDeclineAlertsForUsers(filter bson.D, limit int, after *string, before *string, first *int, last *int) (alerts []*DeclineAlertForUser, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(AlertsCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(AlertsCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		alert := &DeclineAlertForUser{}
		err = cur.Decode(&alert)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		alerts = append(alerts, alert)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return alerts, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// GetDeclineAlertsForProviders gives an array of alerts for providers.
func GetDeclineAlertsForProviders(filter bson.D, limit int, after *string, before *string, first *int, last *int) (alerts []*DeclineAlertForProvider, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(AlertsCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(AlertsCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		alert := &DeclineAlertForProvider{}
		err = cur.Decode(&alert)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		alerts = append(alerts, alert)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return alerts, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateDeclineAlert updates the decline alerts.
func UpdateDeclineAlert(c *DeclineAlert) (*DeclineAlert, error) {
	declineAlert := c
	declineAlert.UpdatedAt = time.Now()
	filter := bson.D{{"_id", declineAlert.ID}}
	db := database.MongoDB
	declineAlertsCollection := db.Collection(AlertsCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := declineAlertsCollection.FindOneAndReplace(context.Background(), filter, declineAlert, findRepOpts).Decode(&declineAlert)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("decline_alert.updated", &declineAlert)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(declineAlert.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return declineAlert, nil
}

// DeleteDeclineAlertByID deletes the decline alert by id.
func DeleteDeclineAlertByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	declineAlertsCollection := db.Collection(AlertsCollection)
	res, err := declineAlertsCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("decline_alert.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (declineAlert *DeclineAlert) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, declineAlert); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (declineAlert *DeclineAlert) MarshalBinary() ([]byte, error) {
	return json.Marshal(declineAlert)
}
