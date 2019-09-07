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

// Currency represents a currency.
type Currency struct {
	ID              primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt       time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt       *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt       time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy       primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Country         string             `json:"country" bson:"country"`
	Name            string             `json:"name" bson:"name"`
	CurrencyCode    string             `json:"currencyCode" bson:"currencyCode"`
	Ratio           string             `json:"ratio" bson:"ratio"`
	ThresholdAmount string             `json:"thresholdAmount" bson:"thresholdAmount"`
	Symbol          string             `json:"symbol" bson:"symbol"`
	IsDefault       bool               `json:"isDefault" bson:"isDefault"`
	IsActive        bool               `json:"isActive" bson:"isActive"`
}

// CreateCurrency creates new currency.
func CreateCurrency(currency Currency) (*Currency, error) {
	currency.CreatedAt = time.Now()
	currency.UpdatedAt = time.Now()
	currency.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(CurrenciesCollection)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := collection.InsertOne(ctx, &currency)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("currency.created", &currency)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(currency.ID.Hex(), currency, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &currency, nil
}

// GetCurrencyByID gives requested currency by id.
func GetCurrencyByID(ID string) (*Currency, error) {
	db := database.MongoDB
	currency := &Currency{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(currency)
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
	err = db.Collection(CurrenciesCollection).FindOne(ctx, filter).Decode(&currency)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, currency, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return currency, nil
}

// GetCurrencies gives an array of currencies.
func GetCurrencies(filter bson.D, limit int, after *string, before *string, first *int, last *int) (currencies []*Currency, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(CurrenciesCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(CurrenciesCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		currency := &Currency{}
		err = cur.Decode(&currency)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		currencies = append(currencies, currency)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return currencies, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateCurrency updates the currency.
func UpdateCurrency(c *Currency) (*Currency, error) {
	currency := c
	currency.UpdatedAt = time.Now()
	filter := bson.D{{"_id", currency.ID}}
	db := database.MongoDB
	currenciesCollection := db.Collection(CurrenciesCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := currenciesCollection.FindOneAndReplace(context.Background(), filter, currency, findRepOpts).Decode(&currency)
	go webhooks.NewWebhookEvent("currency.updated", &currency)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(currency.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return currency, nil
}

// DeleteCurrencyByID deletes currency by id.
func DeleteCurrencyByID(ID string) (bool, error) {
	db := database.MongoDB
	id, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", id}}
	currenciesCollection := db.Collection(CurrenciesCollection)
	res, err := currenciesCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("currency.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (currency *Currency) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, currency); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (currency *Currency) MarshalBinary() ([]byte, error) {
	return json.Marshal(currency)
}
