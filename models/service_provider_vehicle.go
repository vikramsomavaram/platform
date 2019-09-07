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

// ServiceProviderVehicleDetails represents a service provider vehicle.
type ServiceProviderVehicleDetails struct {
	ID                  primitive.ObjectID    `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt           time.Time             `json:"createdAt" bson:"createdAt"`
	DeletedAt           *time.Time            `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt           time.Time             `json:"updatedAt" bson:"updatedAt"`
	CreatedBy           primitive.ObjectID    `json:"createdBy" bson:"createdBy"`
	ServiceProviderID   string                `json:"serviceProviderId" bson:"serviceProviderId"`
	ServiceCompanyID    string                `json:"serviceCompanyId" bson:"serviceCompanyId"`
	VehicleCompanyName  string                `json:"vehicleCompanyName" bson:"vehicleCompanyName"`
	VehicleModelName    string                `json:"vehicleModelName" bson:"vehicleModelName"`
	VehicleYear         string                `json:"vehicleYear" bson:"vehicleYear"`
	VehicleNumber       string                `json:"vehicleNumber" bson:"vehicleNumber"`
	VehicleColor        string                `json:"vehicleColor" bson:"vehicleColor"`
	VehicleImageURL     string                `json:"vehicleImageUrl" bson:"vehicleImageUrl"`
	VehicleLicensePlate string                `json:"vehicleLicensePlate" bson:"vehicleLicensePlate"`
	VehicleType         string                `json:"vehicleType" bson:"vehicleType"`
	EnabledServiceType  []VehicleServiceTypes `json:"enabledServiceType" bson:"enabledServiceType"`
	IsActive            bool                  `json:"isActive" bson:"isActive"`
}

// CreateServiceProviderVehicle creates service provider vehicle.
func CreateServiceProviderVehicle(serviceProviderVehicleDetails ServiceProviderVehicleDetails) (*ServiceProviderVehicleDetails, error) {
	serviceProviderVehicleDetails.CreatedAt = time.Now()
	serviceProviderVehicleDetails.UpdatedAt = time.Now()
	serviceProviderVehicleDetails.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(ServiceProviderVehiclesCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &serviceProviderVehicleDetails)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("service_provider_vehicle.created", &serviceProviderVehicleDetails)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(serviceProviderVehicleDetails.ID.Hex(), serviceProviderVehicleDetails, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &serviceProviderVehicleDetails, nil
}

// GetServiceProviderVehicleByID gives service provider vehicle by id.
func GetServiceProviderVehicleByID(ID string) (*ServiceProviderVehicleDetails, error) {
	db := database.MongoDB
	serviceProviderVehicle := &ServiceProviderVehicleDetails{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(serviceProviderVehicle)
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
	err = db.Collection(ServiceProviderVehiclesCollection).FindOne(ctx, filter).Decode(&serviceProviderVehicle)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, serviceProviderVehicle, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return serviceProviderVehicle, nil
}

// GetServiceProviderVehicles gives a list of service provider vehicles.
func GetServiceProviderVehicles(filter bson.D, limit int, after *string, before *string, first *int, last *int) (serviceProviderVehicles []*ServiceProviderVehicleDetails, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(ServiceProviderVehiclesCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(ServiceProviderVehiclesCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		serviceProviderVehicle := &ServiceProviderVehicleDetails{}
		err = cur.Decode(&serviceProviderVehicle)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		serviceProviderVehicles = append(serviceProviderVehicles, serviceProviderVehicle)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return serviceProviderVehicles, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateServiceProviderVehicle updates service provider vehicle.
func UpdateServiceProviderVehicle(serviceProviderVehicle *ServiceProviderVehicleDetails) (*ServiceProviderVehicleDetails, error) {
	serviceProviderVehicle.UpdatedAt = time.Now()
	filter := bson.D{{"_id", serviceProviderVehicle.ID}}
	db := database.MongoDB
	serviceProviderVehiclesCollection := db.Collection(ServiceProviderVehiclesCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := serviceProviderVehiclesCollection.FindOneAndReplace(context.Background(), filter, serviceProviderVehicle, findRepOpts).Decode(&serviceProviderVehicle)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("service_provider_vehicle.updated", &serviceProviderVehicle)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(serviceProviderVehicle.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return serviceProviderVehicle, nil
}

// DeleteServiceProviderVehicleByID deletes service provider vehicle.
func DeleteServiceProviderVehicleByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	serviceProviderVehiclesCollection := db.Collection(ServiceProviderVehiclesCollection)
	res, err := serviceProviderVehiclesCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("service_provider_vehicle.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (serviceProviderVehicleDetails *ServiceProviderVehicleDetails) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, serviceProviderVehicleDetails); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (serviceProviderVehicleDetails *ServiceProviderVehicleDetails) MarshalBinary() ([]byte, error) {
	return json.Marshal(serviceProviderVehicleDetails)
}
