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

// CancelReason represents a cancel reason.
type CancelReason struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt   *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy   primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Reason      string             `json:"reason" bson:"reason"`
	ServiceType string             `json:"serviceType" bson:"serviceType"`
	Order       string             `json:"order" bson:"order"`
	IsActive    bool               `json:"isActive" bson:"isActive"`
}

// GetCancelReasonByID gives the requested cancel reason by id.
func GetCancelReasonByID(ID string) *CancelReason {
	db := database.MongoDB
	cancelReason := &CancelReason{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(cancelReason)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	id, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		log.Errorln(err)
		return cancelReason
	}
	filter := bson.D{{"_id", id}, {"deletedAt", bson.M{"$exists": false}}}
	ctx := context.Background()
	err = db.Collection(CancelReasonCollection).FindOne(ctx, filter).Decode(&cancelReason)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return cancelReason
		}
		log.Errorln(err)
		return cancelReason
	}
	//set cache item
	err = cacheClient.Set(ID, cancelReason, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return cancelReason
}

// GetCancelReasons gives an array of cancel reasons.
func GetCancelReasons(filter bson.D, limit int, after *string, before *string, first *int, last *int) (cancelReasons []*CancelReason, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(CancelReasonCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(CancelReasonCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
ctx := context.Background()
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		cancelReason := &CancelReason{}
		err = cur.Decode(&cancelReason)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		cancelReasons = append(cancelReasons, cancelReason)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return cancelReasons, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

//UnmarshalBinary required for the redis cache to work
func (cancelReason *CancelReason) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, cancelReason); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (cancelReason *CancelReason) MarshalBinary() ([]byte, error) {
	return json.Marshal(cancelReason)
}

// JobLaterBooking represents a job later booking.
type JobLaterBooking struct {
	ID                          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt                   time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt                   *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt                   time.Time          `json:"updatedAt" bson:"updatedAt"`
	JobType                     string             `json:"jobType" bson:"jobType"`
	BookedBy                    string             `json:"bookedBy" bson:"bookedBy"`
	BookingNumber               int                `json:"bookingNumber" bson:"bookingNumber"`
	Users                       string             `json:"users" bson:"users"`
	Date                        time.Time          `json:"date" bson:"date"`
	ExpectedSourceLocation      Address            `json:"expectedSourceLocation" bson:"expectedSourceLocation"`
	ExpectedDestinationLocation Address            `json:"expectedDestinationLocation" bson:"expectedDestinationLocation"`
	Provider                    string             `json:"provider" bson:"provider"`
	JobDetails                  string             `json:"jobDetails" bson:"jobDetails"`
	Status                      string             `json:"status" bson:"status"`
	IsActive                    bool               `json:"isActive" bson:"isActive"`
}

// GetJobLaterBookingByID gives the requested job later booking by id.
func GetJobLaterBookingByID(ID string) *JobLaterBooking {
	db := database.MongoDB
	jobLaterBooking := &JobLaterBooking{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(jobLaterBooking)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	id, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		log.Errorln(err)
		return jobLaterBooking
	}
	filter := bson.D{{"_id", id}, {"deletedAt", bson.M{"$exists": false}}}
	err = db.Collection(JobLaterBookingCollection).FindOne(context.Background(), filter).Decode(&jobLaterBooking)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return jobLaterBooking
		}
		log.Errorln(err)
		return jobLaterBooking
	}
	//set cache item
	err = cacheClient.Set(ID, jobLaterBooking, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return jobLaterBooking
}

// GetJobLaterBookings gives an array of job later booking.
func GetJobLaterBookings(filter bson.D, limit int, after *string, before *string, first *int, last *int) (jobLaterBookings []*JobLaterBooking, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(JobLaterBookingCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(JobLaterBookingCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
ctx := context.Background()
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		jobLaterBooking := &JobLaterBooking{}
		err = cur.Decode(&jobLaterBooking)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		jobLaterBookings = append(jobLaterBookings, jobLaterBooking)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return jobLaterBookings, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

//UnmarshalBinary required for the redis cache to work
func (jobLaterBooking *JobLaterBooking) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, jobLaterBooking); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (jobLaterBooking *JobLaterBooking) MarshalBinary() ([]byte, error) {
	return json.Marshal(jobLaterBooking)
}

// CreateCancelReason creates new cancel reason.
func CreateCancelReason(cancelReason CancelReason) (*CancelReason, error) {
	cancelReason.CreatedAt = time.Now()
	cancelReason.UpdatedAt = time.Now()
	cancelReason.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(CancelReasonCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &cancelReason)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(cancelReason.ID.Hex(), cancelReason, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("cancel_reason.created", &cancelReason)
	return &cancelReason, nil
}

// UpdateCancelReason updates the reason.
func UpdateCancelReason(c *CancelReason) (*CancelReason, error) {
	cancelReason := c
	cancelReason.UpdatedAt = time.Now()
	filter := bson.D{{"_id", cancelReason.ID}}
	db := database.MongoDB
	companiesCollection := db.Collection(CancelReasonCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := companiesCollection.FindOneAndReplace(context.Background(), filter, cancelReason, findRepOpts).Decode(&cancelReason)
	if err != nil {
		log.Error(err)
	}
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(cancelReason.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("cancel_reason.updated", &cancelReason)
	return cancelReason, nil
}

// DeleteCancelReasonByID deletes the cancel reason by id.
func DeleteCancelReasonByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, _ := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	cancelReasonCollection := db.Collection(CancelReasonCollection)
	res, err := cancelReasonCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
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
	go webhooks.NewWebhookEvent("cancel_reason.updated", &res)
	return true, nil
}
