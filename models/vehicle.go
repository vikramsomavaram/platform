/*
  Copyright (c) 2019. Pandranki Global Private Limited
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

// Vehicle represents a vehicle.
type Vehicle struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy primitive.ObjectID `json:"createdBy" bson:"createdBy"`
}

// ServiceVehicleType represents a service vehicle type.
type ServiceVehicleType struct {
	ID                           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt                    time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt                    time.Time          `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt                    time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy                    primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	VehicleServiceType           VehicleServiceType `json:"vehicleServiceType" bson:"vehicleServiceType"`
	EnablePoolRide               bool               `json:"enablePoolRide" bson:"enablePoolRide"`
	VehicleType                  VehicleType        `json:"vehicleType" bson:"vehicleType"`
	VehicleCategory              VehicleCategory    `json:"vehicleCategory" bson:"vehicleCategory"`
	Location                     string             `json:"location" bson:"location"`
	PricePerKms                  float64            `json:"pricePerKms" bson:"pricePerKms"`
	PricePerMinute               float64            `json:"pricePerMinute" bson:"pricePerMinute"`
	BaseFare                     float64            `json:"baseFare" bson:"baseFare"`
	Commission                   float64            `json:"commission" bson:"commission"`
	MinimumFare                  float64            `json:"minimumFare" bson:"minimumFare"`
	UserCancellationTimeLimit    int                `json:"userCancellationTimeLimit" bson:"userCancellationTimeLimit"`
	UserCancellationCharges      float64            `json:"userCancellationCharges" bson:"userCancellationCharges"`
	WaitingTimeLimit             int                `json:"waitingTimeLimit" bson:"waitingTimeLimit"`
	WaitingCharges               float64            `json:"waitingCharges" bson:"waitingCharges"`
	InTransitWaitingFeePerMinute float64            `json:"inTransitWaitingFeePerMinute" bson:"inTransitWaitingFeePerMinute"`
	PersonCapacity               int                `json:"personCapacity" bson:"personCapacity"`
	PeakTimeSurcharge            bool               `json:"peakTimeSurcharge" bson:"peakTimeSurcharge"`
	NightCharges                 bool               `json:"nightCharges" bson:"nightCharges"`
	VehiclePicture               string             `json:"vehiclePicture" bson:"vehiclePicture"`
	Order                        int                `json:"Order" bson:"Order"`
	IsActive                     bool               `json:"isActive" bson:"isActive"`
}

// CreateServiceVehicleType creates service vehicle type.
func CreateServiceVehicleType(serviceVehicleType ServiceVehicleType) (*ServiceVehicleType, error) {
	serviceVehicleType.CreatedAt = time.Now()
	serviceVehicleType.UpdatedAt = time.Now()
	serviceVehicleType.ID = primitive.NewObjectID()
	db := database.MongoDB
	serviceVehicleTypesCollection := db.Collection(ServiceVehicleTypesCollection)
	ctx := context.Background()
	_, err := serviceVehicleTypesCollection.InsertOne(ctx, &serviceVehicleType)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("service_vehicle_type.created", &serviceVehicleType)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(serviceVehicleType.ID.Hex(), serviceVehicleType, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &serviceVehicleType, nil
}

// GetServiceVehicleTypeByID gives a service vehicle type by id.
func GetServiceVehicleTypeByID(ID string) (*ServiceVehicleType, error) {
	db := database.MongoDB
	serviceVehicleType := &ServiceVehicleType{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(serviceVehicleType)
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
	ctx := context.Background()
	err = db.Collection(ServiceVehicleTypesCollection).FindOne(ctx, filter).Decode(&serviceVehicleType)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, serviceVehicleType, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return serviceVehicleType, nil
}

// GetServiceVehicleTypes gives a list of service vehicle type.
func GetServiceVehicleTypes(filter bson.D, limit int, after *string, before *string, first *int, last *int) (serviceVehicleTypes []*ServiceVehicleType, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(ServiceVehicleTypesCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(ServiceVehicleTypesCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		serviceVehicleType := &ServiceVehicleType{}
		err = cur.Decode(&serviceVehicleType)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		serviceVehicleTypes = append(serviceVehicleTypes, serviceVehicleType)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return serviceVehicleTypes, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateServiceVehicleType updates service vehicle type.
func UpdateServiceVehicleType(s *ServiceVehicleType) (*ServiceVehicleType, error) {
	serviceVehicleType := s
	serviceVehicleType.UpdatedAt = time.Now()
	filter := bson.D{{"_id", serviceVehicleType.ID}}
	db := database.MongoDB
	serviceVehicleTypesCollection := db.Collection(ServiceVehicleTypesCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := serviceVehicleTypesCollection.FindOneAndReplace(context.Background(), filter, serviceVehicleType, findRepOpts).Decode(&serviceVehicleType)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("service_vehicle_type.updated", &serviceVehicleType)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(serviceVehicleType.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return serviceVehicleType, nil
}

// DeleteServiceVehicleTypeByID deletes service vehicle type.
func DeleteServiceVehicleTypeByID(ID string) (bool, error) {
	db := database.MongoDB
	id, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", id}}
	serviceVehicleTypesCollection := db.Collection(ServiceVehicleTypesCollection)
	res, err := serviceVehicleTypesCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	log.Warningln(res)
	if err != nil && res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("service_vehicle_type.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (serviceVehicleType *ServiceVehicleType) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, serviceVehicleType); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (serviceVehicleType *ServiceVehicleType) MarshalBinary() ([]byte, error) {
	return json.Marshal(serviceVehicleType)
}
