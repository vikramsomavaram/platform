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

// VehicleMake represents a vehicle make.
type VehicleMake struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	IsActive  bool               `json:"isActive" bson:"isActive"`
	Make      string             `json:"make" bson:"make"`
}

// CreateVehicleMake creates vehicle make.
func CreateVehicleMake(vehicleMake VehicleMake) (*VehicleMake, error) {
	vehicleMake.CreatedAt = time.Now()
	vehicleMake.UpdatedAt = time.Now()
	vehicleMake.ID = primitive.NewObjectID()
	db := database.MongoDB
	vehicleMakeCollection := db.Collection(VehicleMakeCollection)
	ctx := context.Background()
	_, err := vehicleMakeCollection.InsertOne(ctx, &vehicleMake)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("vehicle_make.created", &vehicleMake)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(vehicleMake.ID.Hex(), vehicleMake, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &vehicleMake, nil
}

// GetVehicleMakeByID gives vehicle make by id.
func GetVehicleMakeByID(ID string) (*VehicleMake, error) {
	db := database.MongoDB
	vehicleMake := &VehicleMake{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(vehicleMake)
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
	err = db.Collection(VehicleMakeCollection).FindOne(ctx, filter).Decode(&vehicleMake)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, vehicleMake, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return vehicleMake, nil
}

// GetVehicleMakes gives a list of vehicle makes.
func GetVehicleMakes(filter bson.D, limit int, after *string, before *string, first *int, last *int) (vehiclesMakes []*VehicleMake, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(VehicleMakeCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(VehicleMakeCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		vehiclesMake := &VehicleMake{}
		err = cur.Decode(&vehiclesMake)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		vehiclesMakes = append(vehiclesMakes, vehiclesMake)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return vehiclesMakes, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateVehicleMake updates vehicle make.
func UpdateVehicleMake(s *VehicleMake) (*VehicleMake, error) {
	vehicleMake := s
	vehicleMake.UpdatedAt = time.Now()
	filter := bson.D{{"_id", vehicleMake.ID}}
	db := database.MongoDB
	vehicleMakeCollection := db.Collection(VehicleMakeCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := vehicleMakeCollection.FindOneAndReplace(context.Background(), filter, vehicleMake, findRepOpts).Decode(&vehicleMake)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("vehicle_make.updated", &vehicleMake)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(vehicleMake.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return vehicleMake, nil
}

// DeleteVehicleMakeByID deletes vehicle make by id.
func DeleteVehicleMakeByID(ID string) (bool, error) {
	db := database.MongoDB
	id, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", id}}
	vehicleMakeCollection := db.Collection(VehicleMakeCollection)
	res, err := vehicleMakeCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	log.Warningln(res)
	if err != nil && res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("vehicle_make.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (vehicleMake *VehicleMake) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, vehicleMake); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (vehicleMake *VehicleMake) MarshalBinary() ([]byte, error) {
	return json.Marshal(vehicleMake)
}
