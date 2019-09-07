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

// Merchant represents a merchant.
type Merchant struct {
	ID                primitive.ObjectID     `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt         time.Time              `json:"createdAt" bson:"createdAt"`
	DeletedAt         *time.Time             `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt         time.Time              `json:"updatedAt" bson:"updatedAt"`
	CreatedBy         string                 `json:"createdBy" bson:"createdBy"`
	MID               string                 `json:"mid" bson:"mid"`
	Zipcode           string                 `json:"zipcode" bson:"zipcode"`
	Email             string                 `json:"email" bson:"email"`
	MobileNo          string                 `json:"mobileNo" bson:"mobileNo"`
	City              string                 `json:"city" bson:"city"`
	State             string                 `json:"state" bson:"state"`
	Country           string                 `json:"country" bson:"country"`
	Currency          string                 `json:"currency" bson:"currency"`
	Name              string                 `json:"name" bson:"name"`
	IsApproved        bool                   `json:"isApproved" bson:"isApproved"`
	IsApprovedOn      time.Time              `json:"isApprovedOn" bson:"isApprovedOn"`
	IsActive          bool                   `json:"isActive" bson:"isActive"`
	IsLocked          bool                   `json:"isLocked" bson:"isLocked"`
	BusinessType      string                 `json:"businessType" bson:"businessType"`
	BillingLabel      string                 `json:"billingLabel" bson:"billingLabel"`
	BusinessModel     string                 `json:"businessModel" bson:"businessModel"`
	WebsiteURL        string                 `json:"websiteUrl" bson:"websiteUrl"`
	GSTNumber         string                 `json:"gstNumber" bson:"gstNumber"`
	CINNumber         string                 `json:"cinNumber" bson:"cinNumber"`
	PANNumber         string                 `json:"panNumber" bson:"panNumber"`
	Address           string                 `json:"address" bson:"address"`
	BankAccount       string                 `json:"bankAccount" bson:"bankAccount"`
	BusinessDocuments []Document             `json:"businessDocuments" bson:"businessDocuments"`
	Users             []string               `json:"users" bson:"users"`
	MCC               string                 `json:"mcc" bson:"mcc"`
	Metadata          map[string]interface{} `json:"metadata" bson:"metadata"`
}

// GetMerchantByMID gives requested merchant details by id.
func (m *Merchant) GetMerchantByMID(mid string) (*Merchant, error) {
	db := database.MongoDB
	merchant := &Merchant{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(mid).Scan(merchant)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	filter := bson.D{{"mid", mid}}
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	err = db.Collection(MerchantsCollection).FindOne(ctx, filter).Decode(&merchant)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(mid, merchant, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return merchant, nil
}

// GetMerchantsByUserID gives a list of merchants.
func GetMerchantsByUserID(uid string) []Merchant {
	db := database.MongoDB
	var merchants []Merchant
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	merchantCollection := db.Collection(MerchantsCollection)
	cur, err := merchantCollection.Find(ctx, bson.D{{"users", uid}})
	if err != nil {
		log.Errorln(err)
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var merchant Merchant
		err := cur.Decode(&merchant)
		if err != nil {
			log.Errorln(err)
			return nil
		}
		merchants = append(merchants, merchant)
	}
	if err := cur.Err(); err != nil {
		log.Errorln(err)
		return nil
	}

	return merchants
}

// CreateMerchant creates merchant.
func CreateMerchant(merchant Merchant) (*Merchant, error) {
	merchant.CreatedAt = time.Now()
	merchant.UpdatedAt = time.Now()
	merchant.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(MerchantsCollection)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := collection.InsertOne(ctx, &merchant)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("merchant.created", &merchant)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(merchant.ID.Hex(), merchant, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &merchant, nil
}

// GetMerchants gives a list of merchants.
func GetMerchants(filter bson.D, limit int, after *string, before *string, first *int, last *int) (merchants []*Merchant, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(MerchantsCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(MerchantsCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		merchant := &Merchant{}
		err = cur.Decode(&merchant)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		merchants = append(merchants, merchant)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return merchants, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateMerchant updates the merchant.
func UpdateMerchant(c *Merchant) (*Merchant, error) {
	merchant := c
	merchant.UpdatedAt = time.Now()
	filter := bson.D{{"_id", merchant.ID}}
	db := database.MongoDB
	companiesCollection := db.Collection(MerchantsCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := companiesCollection.FindOneAndReplace(context.Background(), filter, merchant, findRepOpts).Decode(&merchant)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("merchant.updated", &merchant)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(merchant.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return merchant, nil
}

// DeleteMerchantByID deletes the merchant by id.
func DeleteMerchantByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	merchantsCollection := db.Collection(MerchantsCollection)
	res, err := merchantsCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("merchant.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (merchant *Merchant) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, merchant); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (merchant *Merchant) MarshalBinary() ([]byte, error) {
	return json.Marshal(merchant)
}
