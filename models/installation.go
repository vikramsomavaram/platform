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

// Installation represents a installation.
type Installation struct {
	ID                 primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt          time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt          *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt          time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy          primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	UserID             string             `json:"userId,omitempty" bson:"userId,omitempty"`
	DeviceID           string             `json:"deviceId" bson:"deviceId" validate:"required"`
	FcmToken           string             `bson:"fcmToken" json:"fcmToken" validate:"required"`
	DeviceWidth        float64            `json:"deviceWidth" bson:"deviceWidth"`
	DeviceHeight       float64            `json:"deviceHeight" bson:"deviceHeight"`
	DeviceCountry      string             `json:"deviceCountry,omitempty" bson:"deviceCountry,omitempty"`
	Badge              string             `json:"badge,omitempty" bson:"badge,omitempty"`
	DeviceManufacturer string             `json:"deviceManufacturer,omitempty" bson:"deviceManufacturer,omitempty"`
	SystemVersion      string             `json:"systemVersion,omitempty" bson:"systemVersion,omitempty"`
	AppIdentifier      string             `json:"appIdentifier" validate:"required" bson:"appIdentifier"`
	AppName            string             `json:"appName" bson:"appName" validate:"required"`
	DeviceLocale       string             `json:"deviceLocale,omitempty" bson:"deviceLocale,omitempty"`
	DeviceType         string             `json:"deviceType,omitempty" bson:"deviceType,omitempty"` //android or ios
	Channels           []string           `json:"channels,omitempty" bson:"channels,omitempty"`
	DeviceBrand        string             `json:"deviceBrand,omitempty" bson:"deviceBrand,omitempty"`
	DeviceModel        string             `json:"deviceModel,omitempty" bson:"deviceModel,omitempty"`
	BuildNumber        string             `json:"buildNumber,omitempty" bson:"buildNumber,omitempty"`
	TimeZone           string             `json:"timeZone,omitempty" validate:"required" bson:"timeZone,omitempty"`
	AppVersion         string             `json:"appVersion,omitempty" validate:"required" bson:"appVersion,omitempty"`
}

// CreateInstallation creates installation.
func CreateInstallation(installation *Installation) (*Installation, error) {
	installation.CreatedAt = time.Now()
	installation.UpdatedAt = time.Now()
	db := database.MongoDB
	collection := db.Collection(InstallationsCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &installation)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("installation.created", &installation)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(installation.ID.Hex(), installation, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return installation, nil
}

// GetInstallationByID gives the requested installation using id.
func GetInstallationByID(ID string) (*Installation, error) {
	db := database.MongoDB
	installation := &Installation{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(installation)
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
	err = db.Collection(InstallationsCollection).FindOne(ctx, filter).Decode(&installation)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, installation, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return installation, nil
}

// GetInstallations gives an list of installations.
func GetInstallations(filter bson.D, limit int, after *string, before *string, first *int, last *int) (installations []*Installation, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(InstallationsCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(InstallationsCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
ctx := context.Background()
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		installation := &Installation{}
		err = cur.Decode(&installation)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		installations = append(installations, installation)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return installations, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateInstallation updates installation.
func UpdateInstallation(c *Installation) (*Installation, error) {
	installation := c
	installation.UpdatedAt = time.Now()
	filter := bson.D{{"_id", installation.ID}}
	db := database.MongoDB
	installationsCollection := db.Collection(InstallationsCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := installationsCollection.FindOneAndReplace(context.Background(), filter, installation, findRepOpts).Decode(&installation)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("installation.updated", &installation)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(installation.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return installation, nil
}

// DeleteInstallationByID deletes installation by id.
func DeleteInstallationByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	installationsCollection := db.Collection(InstallationsCollection)
	res, err := installationsCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("installation.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (installation *Installation) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, installation); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (installation *Installation) MarshalBinary() ([]byte, error) {
	return json.Marshal(installation)
}
