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

// RentalPackage represents a rental package.
type RentalPackage struct {
	ID                     primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt              time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt              *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt              time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy              primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Name                   string             `json:"name" bson:"name"`
	RentalTotalPrice       int                `json:"rentalTotalPrice" bson:"rentalTotalPrice"`
	RentalMiles            int                `json:"rentalMiles" bson:"rentalMiles"`
	RentalHour             int                `json:"rentalHour" bson:"rentalHour"`
	AdditionalPricePerMile int                `json:"additionalPricePerMile" bson:"additionalPricePerMile"`
	AdditionalPricePerMin  int                `json:"additionalPricePerMin" bson:"additionalPricePerMin"`
}

// CreateRentalPackage creates a rental package.
func CreateRentalPackage(rentalPackage RentalPackage) (*RentalPackage, error) {
	rentalPackage.CreatedAt = time.Now()
	rentalPackage.UpdatedAt = time.Now()
	rentalPackage.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(RentalPackageCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &rentalPackage)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("rental_package.created", &rentalPackage)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(rentalPackage.ID.Hex(), rentalPackage, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &rentalPackage, nil
}

// GetRentalPackageByID gives a rental package by id.
func GetRentalPackageByID(ID string) (*RentalPackage, error) {
	db := database.MongoDB
	rentalPackage := &RentalPackage{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(rentalPackage)
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
	err = db.Collection(RentalPackageCollection).FindOne(ctx, filter).Decode(&rentalPackage)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, rentalPackage, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return rentalPackage, nil
}

// GetRentalPackages gives a list of rental packages.
func GetRentalPackages(filter bson.D, limit int, after *string, before *string, first *int, last *int) (rentalPacks []*RentalPackage, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(RentalPackageCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(RentalPackageCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
ctx := context.Background()
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		rentalPack := &RentalPackage{}
		err = cur.Decode(&rentalPack)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		rentalPacks = append(rentalPacks, rentalPack)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return rentalPacks, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateRentalPackage updates rental packages.
func UpdateRentalPackage(c *RentalPackage) (*RentalPackage, error) {
	rentalPackage := c
	rentalPackage.UpdatedAt = time.Now()
	filter := bson.D{{"_id", rentalPackage.ID}}
	db := database.MongoDB
	rentalPackageCollection := db.Collection(RentalPackageCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := rentalPackageCollection.FindOneAndReplace(context.Background(), filter, rentalPackage, findRepOpts).Decode(&rentalPackage)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("rental_package.updated", &rentalPackage)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(rentalPackage.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return rentalPackage, nil
}

// DeleteRentalPackageByID deletes rental package by id.
func DeleteRentalPackageByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	rentalPackageCollection := db.Collection(RentalPackageCollection)
	res, err := rentalPackageCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("rental_package.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (rentalPackage *RentalPackage) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, rentalPackage); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (rentalPackage *RentalPackage) MarshalBinary() ([]byte, error) {
	return json.Marshal(rentalPackage)
}
