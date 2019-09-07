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

//PaymentMethod represents a payment method.
type PaymentMethod struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt *time.Time         `json:"-,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`
	Name      string             `json:"name" bson:"name"`
	Type      PaymentMethodType  `json:"type" bson:"type"`
	UserID    string             `json:"userId" bson:"userId"`
}

//CreatePaymentMethod creates payment method.
func CreatePaymentMethod(paymentInstrument PaymentMethod) (*PaymentMethod, error) {
	paymentInstrument.CreatedAt = time.Now()
	paymentInstrument.UpdatedAt = time.Now()
	paymentInstrument.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(UserPaymentMethodsCollection)
	_, err := collection.InsertOne(context.Background(), &paymentInstrument)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("user_payment_method.created", &paymentInstrument)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(paymentInstrument.ID.Hex(), paymentInstrument, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &paymentInstrument, nil
}

// GetPaymentMethodByID gives a payment method by id.
func GetPaymentMethodByID(ID string) (*PaymentMethod, error) {
	db := database.MongoDB
	paymentInstrument := &PaymentMethod{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(paymentInstrument)
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
	err = db.Collection(UserPaymentMethodsCollection).FindOne(context.Background(), filter).Decode(&paymentInstrument)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, paymentInstrument, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return paymentInstrument, nil
}

// GetPaymentMethods gives a list of payment methods.
func GetPaymentMethods(filter bson.D, limit int, after *string, before *string, first *int, last *int) (userPaymentMethods []*PaymentMethod, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(PaymentMethodsCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(PaymentMethodsCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
ctx := context.Background()
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		userPaymentMethod := &PaymentMethod{}
		err = cur.Decode(&userPaymentMethod)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		userPaymentMethods = append(userPaymentMethods, userPaymentMethod)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return userPaymentMethods, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateuserPaymentMethod updates user payment method.
func UpdateuserPaymentMethod(paymentInstrument *PaymentMethod) (*PaymentMethod, error) {
	paymentInstrument.UpdatedAt = time.Now()
	filter := bson.D{{"_id", paymentInstrument.ID}}
	db := database.MongoDB
	userPaymentMethodsCollection := db.Collection(UserPaymentMethodsCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := userPaymentMethodsCollection.FindOneAndReplace(context.Background(), filter, paymentInstrument, findRepOpts).Decode(&paymentInstrument)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("user_payment_method.updated", &paymentInstrument)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(paymentInstrument.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return paymentInstrument, nil
}

// DeleteUserPaymentMethodByID deletes user payment method by id.
func DeleteUserPaymentMethodByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	userPaymentMethodsCollection := db.Collection(UserPaymentMethodsCollection)
	res, err := userPaymentMethodsCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("user_payment_method.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (paymentInstrument *PaymentMethod) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, paymentInstrument); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (paymentInstrument *PaymentMethod) MarshalBinary() ([]byte, error) {
	return json.Marshal(paymentInstrument)
}
