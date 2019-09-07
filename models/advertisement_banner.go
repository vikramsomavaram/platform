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

// AdvertisementBanner represents a advertisement banner.
type AdvertisementBanner struct {
	ID              primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt       time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt       time.Time          `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt       time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy       primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	BannerImage     string             `json:"bannerImage" bson:"bannerImage"`
	BannerName      string             `json:"bannerName" bson:"bannerName"`
	DisplayOrder    int                `json:"displayOrder" bson:"displayOrder"`
	RedirectURL     string             `json:"redirectURL" bson:"redirectURL"`
	TimePeriod      string             `json:"timePeriod" bson:"timePeriod"`
	AddedDate       time.Time          `json:"addedDate" bson:"addedDate"`
	TotalImpression string             `json:"totalImpression" bson:"totalImpression"`
	UsedImpression  string             `json:"usedImpression" bson:"usedImpression"`
	Validity        Validity           `json:"validity" bson:"validity"`
	ClickCount      ClickCount         `json:"clickCount" bson:"clickCount"`
	IsActive        bool               `json:"isActive" bson:"isActive"`
}

// CreateAdvertisementBanner creates new advertisement banner.
func CreateAdvertisementBanner(adBanner AdvertisementBanner) (*AdvertisementBanner, error) {
	adBanner.CreatedAt = time.Now()
	adBanner.UpdatedAt = time.Now()
	adBanner.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(AdvertisementBannersCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &adBanner)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(adBanner.ID.Hex(), adBanner, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("advertisement_banner.created", &adBanner)
	return &adBanner, nil
}

// GetAdvertisementBannerByID gets advertisement banners by ID.
func GetAdvertisementBannerByID(ID string) *AdvertisementBanner {
	db := database.MongoDB
	adBanner := &AdvertisementBanner{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(adBanner)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}

	oID, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		log.Errorln(err)
		return adBanner
	}
	filter := bson.D{{"_id", oID}, {"deletedAt", bson.M{"$exists": false}}}
	ctx := context.Background()
	err = db.Collection(AdvertisementBannersCollection).FindOne(ctx, filter).Decode(&adBanner)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return adBanner
		}
		log.Error(err)
		return adBanner
	}
	//set cache item
	err = cacheClient.Set(ID, adBanner, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return adBanner
}

// GetAdvertisementBanners gets the array of advertisement banners.
func GetAdvertisementBanners(filter bson.D, limit int, after *string, before *string, first *int, last *int) (adBanners []*AdvertisementBanner, totalCount int64, hasPrevious, hasNext bool, err error) {
	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(AdvertisementBannersCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(AdvertisementBannersCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
ctx := context.Background()
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		adBanner := &AdvertisementBanner{}
		err = cur.Decode(&adBanner)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		adBanners = append(adBanners, adBanner)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return adBanners, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateAdvertisementBanner updates the advertisement banners.
func UpdateAdvertisementBanner(c *AdvertisementBanner) (*AdvertisementBanner, error) {
	adBanner := c
	adBanner.UpdatedAt = time.Now()
	filter := bson.D{{"_id", adBanner.ID}}
	db := database.MongoDB
	collection := db.Collection(AdvertisementBannersCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := collection.FindOneAndReplace(context.Background(), filter, adBanner, findRepOpts).Decode(&adBanner)
	if err != nil {
		log.Error(err)
	}
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(adBanner.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("advertisement_banner.updated", &adBanner)
	return adBanner, nil
}

// DeleteAdvertisementBannerByID deletes the advertisement banners by ID.
func DeleteAdvertisementBannerByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	collection := db.Collection(AdvertisementBannersCollection)
	res, err := collection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	log.Warningln(res)
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
	go webhooks.NewWebhookEvent("advertisement_banner.deleted", &res)
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (adBanner *AdvertisementBanner) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, adBanner); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (adBanner *AdvertisementBanner) MarshalBinary() ([]byte, error) {
	return json.Marshal(adBanner)
}
