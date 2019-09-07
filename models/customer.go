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

type Customer struct {
	ID               primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt        time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt        time.Time          `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt        time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy        primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Email            string             `json:"email" bson:"email"`
	FirstName        string             `json:"firstName" bson:"firstName"`
	LastName         string             `json:"lastName" bson:"lastName"`
	Role             string             `json:"role" bson:"role"`
	Username         string             `json:"username" bson:"username"`
	Billing          Billing            `json:"billing" bson:"billing"`
	Shipping         Shipping           `json:"shipping" bson:"shipping"`
	IsPayingCustomer bool               `json:"isPayingCustomer" bson:"isPayingCustomer"`
	AvatarURL        string             `json:"avatarURL" bson:"avatarURL"`
	MetaData         MetaData           `json:"metaData" bson:"metaData"`
	IsActive         bool               `json:"isActive" bson:"isActive"`
}

// CreateCustomer creates new customer.
func CreateCustomer(customer Customer) (*Customer, error) {
	customer.CreatedAt = time.Now()
	customer.UpdatedAt = time.Now()
	customer.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(CustomersCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &customer)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("customer.created", &customer)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(customer.ID.Hex(), customer, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &customer, nil
}

// GetCustomerByID gives customer by id.
func GetCustomerByID(ID string) (*Customer, error) {
	db := database.MongoDB
	customer := &Customer{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(customer)
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
	err = db.Collection(CustomersCollection).FindOne(ctx, filter).Decode(&customer)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, customer, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return customer, nil
}

// GetCustomers gives a list of customers.
func GetCustomers(filter bson.D, limit int, after *string, before *string, first *int, last *int) (customers []*Customer, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(CustomersCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(CustomersCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		customer := &Customer{}
		err = cur.Decode(&customer)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		customers = append(customers, customer)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return customers, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateCustomer updates customer.
func UpdateCustomer(s *Customer) (*Customer, error) {
	customer := s
	customer.UpdatedAt = time.Now()
	filter := bson.D{{"_id", customer.ID}}
	db := database.MongoDB
	customersCollection := db.Collection(CustomersCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := customersCollection.FindOneAndReplace(context.Background(), filter, customer, findRepOpts).Decode(&customer)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("customer.updated", &customer)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(customer.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return customer, nil
}

// DeleteCustomerByID deletes customer by id.
func DeleteCustomerByID(ID string) (bool, error) {
	db := database.MongoDB
	id, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", id}}
	customersCollection := db.Collection(CustomersCollection)
	res, err := customersCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil && res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("customer.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (customer *Customer) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, customer); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (customer *Customer) MarshalBinary() ([]byte, error) {
	return json.Marshal(customer)
}
