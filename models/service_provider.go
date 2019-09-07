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

// ServiceProvider represents a service provider.
type ServiceProvider struct {
	ID                 primitive.ObjectID     `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt          time.Time              `json:"createdAt" bson:"createdAt"`
	DeletedAt          *time.Time             `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt          time.Time              `json:"updatedAt" bson:"updatedAt"`
	CreatedBy          primitive.ObjectID     `json:"createdBy" bson:"createdBy"`
	Blocked            bool                   `json:"blocked" bson:"blocked"`
	User               primitive.ObjectID     `json:"user" bson:"user"`
	CompanyID          primitive.ObjectID     `json:"companyId" bson:"companyId"`
	BankDetails        primitive.ObjectID     `json:"bankDetails" bson:"bankDetails"`
	Metadata           map[string]interface{} `json:"metadata" bson:"metadata"`
	FirstName          string                 `json:"firstName" bson:"firstName"`
	LastName           string                 `json:"lastName" bson:"lastName"`
	Email              string                 `json:"email" bson:"email"`
	Password           string                 `json:"password" bson:"password"`
	Gender             Gender                 `json:"gender" bson:"gender"`
	ProfilePicture     string                 `json:"profilePicture" bson:"profilePicture"`
	Country            string                 `json:"country" bson:"country"`
	State              string                 `json:"state" bson:"state"`
	City               string                 `json:"city" bson:"city"`
	Address            Address                `json:"address" bson:"address"`
	ZipCode            int                    `json:"zipCode" bson:"zipCode"`
	MobileNumber       string                 `json:"mobileNumber" bson:"mobileNumber"`
	Language           string                 `json:"language" bson:"language"`
	Currency           string                 `json:"currency" bson:"currency"`
	ServiceCategory    []ServiceCategory      `json:"serviceCategory" bson:"serviceCategory"`
	ServiceSubCategory []primitive.ObjectID   `json:"serviceSubCategory" bson:"serviceSubCategory"`
	ApprovedAt         *time.Time             `json:"approvedAt" bson:"approvedAt"`
	ApprovedBy         *primitive.ObjectID    `json:"approvedBy" bson:"approvedBy"`
	IsActive           bool                   `json:"isActive" bson:"isActive"`
	//RazorPay Account ID is stored in metadata key "razorpay_route_account_id" same goes for
}

// CreateServiceProvider creates new service provider.
func CreateServiceProvider(serviceProvider *ServiceProvider) (*ServiceProvider, error) {
	serviceProvider.CreatedAt = time.Now()
	serviceProvider.UpdatedAt = time.Now()
	serviceProvider.ID = primitive.NewObjectID()
	db := database.MongoDB
	providersCollections := db.Collection(ServiceProvidersCollection)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := providersCollections.InsertOne(ctx, &serviceProvider)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("service_provider.created", &serviceProvider)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(serviceProvider.ID.Hex(), serviceProvider, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return serviceProvider, nil
}

// GetServiceProviderByID gives  service provider by id.
func GetServiceProviderByID(ID string) *ServiceProvider {
	db := database.MongoDB
	serviceProvider := &ServiceProvider{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(serviceProvider)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}

	id, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		log.Errorln(err)
		return serviceProvider
	}
	filter := bson.D{{"_id", id}, {"deletedAt", bson.M{"$exists": false}}}
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	err = db.Collection(ServiceProvidersCollection).FindOne(ctx, filter).Decode(&serviceProvider)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return serviceProvider
		}
		log.Errorln(err)
		return serviceProvider
	}
	//set cache item
	err = cacheClient.Set(ID, serviceProvider, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return serviceProvider
}

// GetServiceProviderByFilter gives  service provider by filter.
func GetServiceProviderByFilter(filter bson.D) *ServiceProvider {
	db := database.MongoDB
	serviceProvider := &ServiceProvider{}
	err, filterHash := genBsonHash(filter)
	if err != nil {
		log.Errorln(err)
	}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err = cacheClient.Get(filterHash).Scan(serviceProvider)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}

	filter = append(filter, bson.E{"deletedAt", bson.M{"$exists": false}})
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	err = db.Collection(ServiceProvidersCollection).FindOne(ctx, filter).Decode(&serviceProvider)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return serviceProvider
		}
		log.Errorln(err)
		return serviceProvider
	}
	//set cache item
	err = cacheClient.Set(filterHash, serviceProvider, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return serviceProvider
}

// GetServiceProviders gives a list of service provider.
func GetServiceProviders(filter bson.D, limit int, after *string, before *string, first *int, last *int) (serviceProviders []*ServiceProvider, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(ServiceProvidersCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(ServiceProvidersCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		serviceProvider := &ServiceProvider{}
		err = cur.Decode(&serviceProvider)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		serviceProviders = append(serviceProviders, serviceProvider)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return serviceProviders, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateServiceProvider updates service provider.
func UpdateServiceProvider(s *ServiceProvider) (*ServiceProvider, error) {
	serviceProvider := s
	serviceProvider.UpdatedAt = time.Now()
	filter := bson.D{{"_id", serviceProvider.ID}}
	db := database.MongoDB
	serviceProvidersCollection := db.Collection(ServiceProvidersCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := serviceProvidersCollection.FindOneAndReplace(context.Background(), filter, serviceProvider, findRepOpts).Decode(&serviceProvider)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("service_provider.updated", &serviceProvider)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(serviceProvider.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return serviceProvider, nil
}

// DeleteServiceProviderByID deletes service provider by id.
func DeleteServiceProviderByID(ID string) (bool, error) {
	db := database.MongoDB
	id, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", id}}
	serviceProvidersCollection := db.Collection(ServiceProvidersCollection)
	res, err := serviceProvidersCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	log.Warningln(res)
	if err != nil && res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("service_provider.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (serviceProvider *ServiceProvider) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, serviceProvider); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (serviceProvider *ServiceProvider) MarshalBinary() ([]byte, error) {
	return json.Marshal(serviceProvider)
}
