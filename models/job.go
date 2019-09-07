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

// Job represents a job.
type Job struct {
	ID                  primitive.ObjectID    `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt           time.Time             `json:"createdAt" bson:"createdAt"`
	DeletedAt           *time.Time            `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt           time.Time             `json:"updatedAt" bson:"updatedAt"`
	CreatedBy           primitive.ObjectID    `json:"createdBy" bson:"createdBy"`
	CancelledAt         *time.Time            `json:"cancelledAt" bson:"cancelledAt"`
	JobType             ServiceCategory       `json:"jobType" bson:"jobType"`
	BookedFor           string                `json:"bookedFor" bson:"bookedFor"`
	BookingNumber       string                `json:"bookingNumber" bson:"bookingNumber"`
	ToAddress           Address               `json:"toAddress" bson:"toAddress"`
	FromAddress         Address               `json:"fromAddress" bson:"fromAddress"`
	JobDate             time.Time             `json:"jobDate" bson:"jobDate"`
	CompanyID           string                `json:"companyId" bson:"companyId"`
	ProviderID          string                `json:"providerId" bson:"providerId"`
	UserID              string                `json:"userId" bson:"userId"`
	ServiceVehicleID    *primitive.ObjectID   `bson:"serviceVehicleID"`
	FareAmount          float64               `json:"fareAmount" bson:"fareAmount"`
	EstimatedFareAmount float64               `json:"estimatedFareAmount" bson:"estimatedFareAmount"`
	ServiceType         string                `json:"serviceType" bson:"serviceType"`
	ServiceOrderItems   *[]*ServiceOrderInput `json:"serviceOrderItems" bson:"serviceOrderItems"`
	InvoiceID           string                `json:"invoiceId" bson:"invoiceId"`
}

// CreateJob creates new job.
func CreateJob(job *Job) (*Job, error) {
	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()
	db := database.MongoDB
	collection := db.Collection(JobsCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &job)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(job.ID.Hex(), job, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("job.created", &job)
	return job, nil
}

// GetJobByID gives the requested job by id.
func GetJobByID(ID string) (*Job, error) {
	db := database.MongoDB
	job := &Job{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(job)
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
	err = db.Collection(JobsCollection).FindOne(ctx, filter).Decode(&job)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, job, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return job, nil
}

// GetJobs gives a list of jobs.
func GetJobs(filter bson.D, limit int, after *string, before *string, first *int, last *int) (jobs []*Job, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(JobsCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(JobsCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
ctx := context.Background()
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		job := &Job{}
		err = cur.Decode(&job)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		jobs = append(jobs, job)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return jobs, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

//UnmarshalBinary required for the redis cache to work
func (job *Job) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, job); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (job *Job) MarshalBinary() ([]byte, error) {
	return json.Marshal(job)
}
