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

//Service represents a service.
type Service struct {
	ID                        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt                 time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt                 *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt                 time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy                 primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Name                      string             `json:"name" bson:"name"`
	PriceBasedOn              PriceBasedOn       `json:"priceBasedOn" bson:"priceBasedOn"`
	CommissionOnMaterial      bool               `json:"commissionOnMaterial" bson:"commissionOnMaterial"`
	Category                  ServiceCategory    `json:"category" bson:"category"`
	UserCancellationTimeLimit int                `json:"userCancellationTimeLimit" bson:"userCancellationTimeLimit"`
	UserCancellationCharges   float64            `json:"userCancellationCharges" bson:"userCancellationCharges"`
	WaitingTimeLimit          int                `json:"waitingTimeLimit" bson:"waitingTimeLimit"`
	WaitingCharges            float64            `json:"waitingCharges" bson:"waitingCharges"`
	CategoryViewType          CategoryViewType   `json:"categoryViewType" bson:"categoryViewType"`
	DisplayOrder              int                `json:"displayOrder" bson:"displayOrder"`
	Icon                      string             `json:"icon" bson:"icon"`
	Tags                      []string           `json:"tags" bson:"tags"`
	IsActive                  bool               `json:"isActive" bson:"isActive"`
}

// CreateService creates new service.
func CreateService(service Service) (*Service, error) {
	service.CreatedAt = time.Now()
	service.UpdatedAt = time.Now()
	db := database.MongoDB
	collection := db.Collection(ServicesCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &service)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("services.created", &service)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(service.ID.Hex(), service, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &service, nil
}

// GetServiceByID gives service by id.
func GetServiceByID(ID string) *Service {
	db := database.MongoDB
	service := &Service{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(service)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}

	id, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		log.Errorln(err)
		return service
	}
	filter := bson.D{{"_id", id}, {"deletedAt", bson.M{"$exists": false}}}
	ctx := context.Background()
	err = db.Collection(ServicesCollection).FindOne(ctx, filter).Decode(&service)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return service
		}
		log.Errorln(err)
		return service
	}
	//set cache item
	err = cacheClient.Set(ID, service, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return service
}

// GetServices gives a list of services.
func GetServices(filter bson.D, limit int, after *string, before *string, first *int, last *int) (services []*Service, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(ServicesCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(ServicesCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
ctx := context.Background()
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		service := &Service{}
		err = cur.Decode(&service)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		services = append(services, service)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return services, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateService updates service.
func UpdateService(s *Service) (*Service, error) {
	service := s
	service.UpdatedAt = time.Now()
	filter := bson.D{{"_id", service.ID}}
	db := database.MongoDB
	servicesCollection := db.Collection(ServicesCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := servicesCollection.FindOneAndReplace(context.Background(), filter, service, findRepOpts).Decode(&service)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("services.updated", &service)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(service.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return service, nil
}

// DeleteServiceByID deletes service by id.
func DeleteServiceByID(ID string) (bool, error) {
	db := database.MongoDB
	id, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", id}}
	servicesCollection := db.Collection(ServicesCollection)
	res, err := servicesCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil && res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("services.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (service *Service) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, service)
}

//MarshalBinary required for the redis cache to work
func (service *Service) MarshalBinary() ([]byte, error) {
	return json.Marshal(service)
}
