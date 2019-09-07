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

//Address user address with latitude and longitude
type Address struct {
	ID                 primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt          time.Time          `json:"createdAt" bson:"createdAt" `
	DeletedAt          *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt          time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy          primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Name               string             `json:"name" bson:"name"`
	FirstName          string             `json:"firstName" bson:"firstName"`
	LastName           string             `json:"lastName" bson:"lastName"`
	CompanyName        string             `json:"companyName" bson:"companyName"`
	City               City               `json:"city" bson:"city"`
	State              State              `json:"state" bson:"state"`
	Country            Country            `json:"country" bson:"country"`
	PostCode           int                `json:"postCode" bson:"postCode"`
	AddressDescription string             `json:"addressDescription" bson:"addressDescription"`
	Latitude           float64            `json:"latitude" bson:"latitude"`
	Longitute          float64            `json:"longitute" bson:"longitude"`
}

// CreateAddress creates new address.
func CreateAddress(address Address) (*Address, error) {
	address.CreatedAt = time.Now()
	address.UpdatedAt = time.Now()
	address.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(AddressesCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &address)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(address.ID.Hex(), address, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("address.created", &address)
	return &address, nil
}

// GetAddressByID gets address by ID.
func GetAddressByID(ID string) *Address {
	db := database.MongoDB
	address := &Address{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(address)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	oID, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		log.Errorln(err)
		return address
	}
	filter := bson.D{{"_id", oID}, {"deletedAt", bson.M{"$exists": false}}}
	ctx := context.Background()
	err = db.Collection(AddressesCollection).FindOne(ctx, filter).Decode(&address)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return address
		}
		log.Errorln(err)
		return address
	}
	//set cache item
	err = cacheClient.Set(ID, address, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return address
}

// GetAddresses gets the array of addresses.
func GetAddresses(filter bson.D, limit int, after *string, before *string, first *int, last *int) (addresses []*Address, totalCount int64, hasPrevious, hasNext bool, err error) {
	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(AddressesCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(AddressesCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		address := &Address{}
		err = cur.Decode(&address)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		addresses = append(addresses, address)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return addresses, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateAddress updates the address.
func UpdateAddress(address *Address) (*Address, error) {
	address.UpdatedAt = time.Now()
	filter := bson.D{{"_id", address.ID}}
	db := database.MongoDB
	collection := db.Collection(AddressesCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := collection.FindOneAndReplace(context.Background(), filter, address, findRepOpts).Decode(&address)
	if err != nil {
		log.Error(err)
	}
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(address.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("address.updated", &address)
	return address, nil
}

// DeleteAddressByID deletes the address by ID.
func DeleteAddressByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	collection := db.Collection(AddressesCollection)
	res, err := collection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
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
	go webhooks.NewWebhookEvent("address.deleted", &res)
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (address *Address) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, address); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (address *Address) MarshalBinary() ([]byte, error) {
	return json.Marshal(address)
}
