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

// AppVersion represents a app version.
type AppVersion struct {
	ID             primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt      time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt      *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt      time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy      primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	MinimumVersion string             `json:"minimumVersion" bson:"minimumVersion"`
	LatestVersion  string             `json:"latestVersion" bson:"latestVersion"`
	DownloadURL    string             `json:"downloadUrl" bson:"downloadUrl"`
	Channel        string             `json:"channel" bson:"channel"`
	IsActive       bool               `json:"isActive" bson:"isActive"`
}

// CreateAppVersion is used to create new app versions.
func CreateAppVersion(appVersion AppVersion) (*AppVersion, error) {
	appVersion.CreatedAt = time.Now()
	appVersion.UpdatedAt = time.Now()
	appVersion.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(AppVersionsCollection)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := collection.InsertOne(ctx, &appVersion)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(appVersion.ID.Hex(), appVersion, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("app_version.created", &appVersion)
	return &appVersion, nil
}

// GetAppVersionByID returns requested app version by id.
func GetAppVersionByID(ID string) *AppVersion {
	db := database.MongoDB
	appVersion := &AppVersion{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(appVersion)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	oID, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		log.Errorln(err)
		return appVersion
	}
	filter := bson.D{{"_id", oID}, {"deletedAt", bson.M{"$exists": false}}}
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	err = db.Collection(AppVersionsCollection).FindOne(ctx, filter).Decode(&appVersion)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return appVersion
		}
		log.Errorln(err)
		return appVersion
	}
	//set cache item
	err = cacheClient.Set(ID, appVersion, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return appVersion
}

// GetAppLatestVersion returns the latest app version.
func GetAppLatestVersion() (*AppVersion, error) {
	db := database.MongoDB
	appVersion := &AppVersion{}
	filter := bson.D{{"deletedAt", bson.M{"$exists": false}}}
	err := db.Collection(AppVersionsCollection).FindOne(context.Background(), filter).Decode(&appVersion)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	return appVersion, nil
}

// GetAppVersions returns a list of app versions.
func GetAppVersions(filter bson.D, limit int, after *string, before *string, first *int, last *int) (appVersions []*AppVersion, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(AppVersionsCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(AppVersionsCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		appVersion := &AppVersion{}
		err = cur.Decode(&appVersion)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		appVersions = append(appVersions, appVersion)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return appVersions, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil

}

// UpdateAppVersion updates the current app version.
func UpdateAppVersion(c *AppVersion) (*AppVersion, error) {
	appVersion := c
	appVersion.UpdatedAt = time.Now()
	filter := bson.D{{"_id", appVersion.ID}}
	db := database.MongoDB
	companiesCollection := db.Collection(AppVersionsCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := companiesCollection.FindOneAndReplace(context.Background(), filter, appVersion, findRepOpts).Decode(&appVersion)
	if err != nil {
		log.Error(err)
	}
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(appVersion.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("app_version.updated", &appVersion)
	return appVersion, nil
}

// DeleteAppVersionByID deletes app version by id.
func DeleteAppVersionByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	companiesCollection := db.Collection(AppVersionsCollection)
	res, err := companiesCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("app_version.deleted", &res)
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (appVersion *AppVersion) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, appVersion); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (appVersion *AppVersion) MarshalBinary() ([]byte, error) {
	return json.Marshal(appVersion)
}
