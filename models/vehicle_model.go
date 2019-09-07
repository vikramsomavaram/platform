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

// VehicleModel represents a vehicle model.
type VehicleModel struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	IsActive  bool               `json:"isActive" bson:"isActive"`
	Model     string             `json:"model" bson:"model"`
	Make      string             `json:"make" bson:"make"`
}

// CreateVehicleModel creates vehicle model.
func CreateVehicleModel(vehicleModel VehicleModel) (*VehicleModel, error) {
	vehicleModel.CreatedAt = time.Now()
	vehicleModel.UpdatedAt = time.Now()
	vehicleModel.ID = primitive.NewObjectID()
	db := database.MongoDB
	vehicleModelsCollection := db.Collection(VehicleModelsCollection)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := vehicleModelsCollection.InsertOne(ctx, &vehicleModel)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("vehicle_model.created", &vehicleModel)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(vehicleModel.ID.Hex(), vehicleModel, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &vehicleModel, nil
}

// GetVehicleModelByID gives a vehicle model by id.
func GetVehicleModelByID(ID string) (*VehicleModel, error) {
	db := database.MongoDB
	vehicleModel := &VehicleModel{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(vehicleModel)
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
	err = db.Collection(VehicleModelsCollection).FindOne(ctx, filter).Decode(&vehicleModel)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, vehicleModel, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return vehicleModel, nil
}

// GetVehicleModels gives a list of vehicle models.
func GetVehicleModels(filter bson.D, limit int, after *string, before *string, first *int, last *int) (vehicleModels []*VehicleModel, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(VehicleModelsCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(VehicleModelsCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		vehicleModel := &VehicleModel{}
		err = cur.Decode(&vehicleModel)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		vehicleModels = append(vehicleModels, vehicleModel)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return vehicleModels, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateVehicleModel updates vehicle model.
func UpdateVehicleModel(s *VehicleModel) (*VehicleModel, error) {
	vehicleModel := s
	vehicleModel.UpdatedAt = time.Now()
	filter := bson.D{{"_id", vehicleModel.ID}}
	db := database.MongoDB
	vehicleModelsCollection := db.Collection(VehicleModelsCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := vehicleModelsCollection.FindOneAndReplace(context.Background(), filter, vehicleModel, findRepOpts).Decode(&vehicleModel)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("vehicle_model.updated", &vehicleModel)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(vehicleModel.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return vehicleModel, nil
}

// DeleteVehicleModelByID deletes vehicle model by id.
func DeleteVehicleModelByID(ID string) (bool, error) {
	db := database.MongoDB
	id, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", id}}
	vehicleModelsCollection := db.Collection(VehicleModelsCollection)
	res, err := vehicleModelsCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	log.Warningln(res)
	if err != nil && res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("vehicle_model.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (vehicleModel *VehicleModel) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, vehicleModel); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (vehicleModel *VehicleModel) MarshalBinary() ([]byte, error) {
	return json.Marshal(vehicleModel)
}
