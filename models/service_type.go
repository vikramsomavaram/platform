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

// ServiceType represents a service type.
type ServiceType struct {
	ID                   primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	ServiceSubCategoryID string             `json:"serviceSubCategoryId" bson:"serviceSubCategoryId"`
	CreatedAt            time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt            *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt            time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy            primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	ServiceType          string             `json:"serviceType" bson:"serviceType"`
	DisplayOrder         int                `json:"displayOrder" bson:"displayOrder"`
	Location             string             `json:"location" bson:"location"`
	IsActive             bool               `json:"isActive" bson:"isActive"`
	ServiceCategory      string             `json:"serviceCategory" bson:"serviceCategory"`
	ServiceDescription   string             `json:"serviceDescription" bson:"serviceDescription"`
	FareType             FareType           `json:"fareType" bson:"fareType"`
	ServiceCharge        float64            `json:"serviceCharge" bson:"serviceCharge"`
	Commission           float64            `json:"commission" bson:"commission"`
	AllowQuantity        bool               `json:"allowQuantity" bson:"allowQuantity"`
}

// CreateServiceType creates service type.
func CreateServiceType(serviceType ServiceType) (*ServiceType, error) {
	serviceType.CreatedAt = time.Now()
	serviceType.UpdatedAt = time.Now()
	serviceType.ID = primitive.NewObjectID()
	db := database.MongoDB
	installationCollection := db.Collection(ServiceTypesCollection)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := installationCollection.InsertOne(ctx, &serviceType)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("service_type.created", &serviceType)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(serviceType.ID.Hex(), serviceType, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &serviceType, nil
}

// GetServiceTypeByID gives service type by id.
func GetServiceTypeByID(ID string) (*ServiceType, error) {
	db := database.MongoDB
	serviceType := &ServiceType{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(serviceType)
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
	err = db.Collection(ServiceTypesCollection).FindOne(ctx, filter).Decode(&serviceType)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, serviceType, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return serviceType, nil
}

// GetServiceTypes gives a list of service types.
func GetServiceTypes(filter bson.D, limit int, after *string, before *string, first *int, last *int) (serviceTypes []*ServiceType, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(ServiceTypesCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(ServiceTypesCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		serviceType := &ServiceType{}
		err = cur.Decode(&serviceType)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		serviceTypes = append(serviceTypes, serviceType)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return serviceTypes, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateServiceType updates service type.
func UpdateServiceType(s *ServiceType) (*ServiceType, error) {
	serviceType := s
	serviceType.UpdatedAt = time.Now()
	filter := bson.D{{"_id", serviceType.ID}}
	db := database.MongoDB
	serviceTypesCollection := db.Collection(ServiceTypesCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := serviceTypesCollection.FindOneAndReplace(context.Background(), filter, serviceType, findRepOpts).Decode(&serviceType)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("service_type.updated", &serviceType)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(serviceType.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return serviceType, nil
}

// DeleteServiceTypeByID deletes service type by id.
func DeleteServiceTypeByID(ID string) (bool, error) {
	db := database.MongoDB
	id, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", id}}
	serviceTypesCollection := db.Collection(ServiceTypesCollection)
	res, err := serviceTypesCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil && res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("service_type.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (serviceType *ServiceType) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, serviceType); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (serviceType *ServiceType) MarshalBinary() ([]byte, error) {
	return json.Marshal(serviceType)
}
