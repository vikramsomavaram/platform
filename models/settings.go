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
	"time"
)

// MarketSettings represents a market settings.
type MarketSettings struct {
	ID           primitive.ObjectID  `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt    time.Time           `json:"createdAt" bson:"createdAt"`
	DeletedAt    *time.Time          `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt    time.Time           `json:"updatedAt" bson:"updatedAt"`
	General      GeneralSetting      `json:"general" bson:"general"`
	Email        EmailSetting        `json:"email" bson:"email"`
	Appearance   AppearanceSetting   `json:"appearance" bson:"appearance"`
	Sms          SMSSetting          `json:"sms" bson:"sms"`
	SocialMedia  SocialMediaSetting  `json:"socialMedia" bson:"socialMedia"`
	App          AppSetting          `json:"app" bson:"app"`
	Installation InstallationSetting `json:"installation" bson:"installation"`
	Store        StoreSetting        `json:"store" bson:"store"`
	Payment      PaymentSetting      `json:"payment" bson:"payment"`
	IsActive     bool                `json:"isActive" bson:"isActive"`
}

// CreateMarketSettings creates market settings.
func CreateMarketSettings(setting *MarketSettings) (*MarketSettings, error) {
	setting.CreatedAt = time.Now()
	setting.UpdatedAt = time.Now()
	setting.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(MarketSettingsCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &setting)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("market_settings.created", &setting)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(setting.ID.Hex(), setting, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return setting, nil
}

// GetMarketSettingsByID gives market settings by id.
func GetMarketSettingsByID(ID string) (*MarketSettings, error) {
	db := database.MongoDB
	setting := &MarketSettings{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(setting)
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
	err = db.Collection(MarketSettingsCollection).FindOne(ctx, filter).Decode(&setting)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, setting, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return setting, nil
}

// GetMarketSettings gives a list of market settings.
func GetMarketSettings(filter bson.D, limit int, after *string, before *string, first *int, last *int) (marketSettings []*MarketSettings, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(MarketSettingsCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(MarketSettingsCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
ctx := context.Background()
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		marketSetting := &MarketSettings{}
		err = cur.Decode(&marketSetting)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		marketSettings = append(marketSettings, marketSetting)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return marketSettings, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

//UnmarshalBinary required for the redis cache to work
func (setting *MarketSettings) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, setting); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (setting *MarketSettings) MarshalBinary() ([]byte, error) {
	return json.Marshal(setting)
}

// SEOSetting represents a seo settings.
type SEOSetting struct {
	ID              primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt       time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt       *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt       time.Time          `json:"updatedAt" bson:"updatedAt"`
	PageName        string             `json:"pageName" bson:"pageName"`
	PageTitle       string             `json:"pageTitle" bson:"pageTitle"`
	MetaKeyword     string             `json:"metaKeyword" bson:"metaKeyword"`
	MetaDescription string             `json:"metaDescription" bson:"metaDescription"`
	IsActive        bool               `json:"isActive" bson:"isActive"`
}

// CreateSEOSetting creates new seo settings.
func CreateSEOSetting(seoSetting *SEOSetting) (*SEOSetting, error) {
	seoSetting.CreatedAt = time.Now()
	seoSetting.UpdatedAt = time.Now()
	seoSetting.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(SEOSettingsCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &seoSetting)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("seo_settings.created", &seoSetting)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(seoSetting.ID.Hex(), seoSetting, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return seoSetting, nil
}

// GetSEOSettingByID gives seo settigns by id.
func GetSEOSettingByID(ID string) (*SEOSetting, error) {
	db := database.MongoDB
	seoSetting := &SEOSetting{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(seoSetting)
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
	err = db.Collection(SEOSettingsCollection).FindOne(ctx, filter).Decode(&seoSetting)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, seoSetting, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return seoSetting, nil
}

// GetSEOSettings gives a list of seo settings.
func GetSEOSettings(filter bson.D, limit int, after *string, before *string, first *int, last *int) (seoSettings []*SEOSetting, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(SEOSettingsCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(SEOSettingsCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
ctx := context.Background()
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		seoSetting := &SEOSetting{}
		err = cur.Decode(&seoSetting)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		seoSettings = append(seoSettings, seoSetting)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return seoSettings, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

//UnmarshalBinary required for the redis cache to work
func (seoSetting *SEOSetting) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, seoSetting); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (seoSetting *SEOSetting) MarshalBinary() ([]byte, error) {
	return json.Marshal(seoSetting)
}
