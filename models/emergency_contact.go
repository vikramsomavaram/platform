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

// EmergencyContact represents a emergency contact.
type EmergencyContact struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt" `
	DeletedAt *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Name      string             `json:"name" bson:"name"`
	MobileNo  int                `json:"mobileNo" bson:"mobileNo"`
	EmailID   string             `json:"emailID" bson:"emailID"`
}

// CreateEmergencyContact creates new emergency contact.
func CreateEmergencyContact(emergencyContact EmergencyContact) (*EmergencyContact, error) {
	emergencyContact.CreatedAt = time.Now()
	emergencyContact.UpdatedAt = time.Now()
	emergencyContact.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(EmergencyContactsCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &emergencyContact)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("emergency_contact.created", &emergencyContact)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(emergencyContact.ID.Hex(), emergencyContact, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &emergencyContact, nil
}

// GetEmergencyContactByID gets emergency Contact by ID.
func GetEmergencyContactByID(ID string) (*EmergencyContact, error) {
	db := database.MongoDB
	emergencyContact := &EmergencyContact{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(emergencyContact)
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
	err = db.Collection(EmergencyContactsCollection).FindOne(ctx, filter).Decode(&emergencyContact)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, emergencyContact, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return emergencyContact, nil
}

// GetEmergencyContacts gets the array of emergency contacts.
func GetEmergencyContacts(filter bson.D, limit int, after *string, before *string, first *int, last *int) (emergencyContacts []*EmergencyContact, totalCount int64, hasPrevious, hasNext bool, err error) {
	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(EmergencyContactsCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(EmergencyContactsCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
ctx := context.Background()
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		emergencyContact := &EmergencyContact{}
		err = cur.Decode(&emergencyContact)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		emergencyContacts = append(emergencyContacts, emergencyContact)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return emergencyContacts, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateEmergencyContact updates the emergency contact.
func UpdateEmergencyContact(emergencyContact *EmergencyContact) (*EmergencyContact, error) {
	emergencyContact.UpdatedAt = time.Now()
	filter := bson.D{{"_id", emergencyContact.ID}}
	db := database.MongoDB
	collection := db.Collection(EmergencyContactsCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := collection.FindOneAndReplace(context.Background(), filter, emergencyContact, findRepOpts).Decode(&emergencyContact)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("emergency_contact.updated", &emergencyContact)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(emergencyContact.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return emergencyContact, nil
}

// DeleteEmergencyContactByID deletes the emergency contact by ID.
func DeleteEmergencyContactByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	collection := db.Collection(EmergencyContactsCollection)
	res, err := collection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("emergency_contact.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (emergencyContact *EmergencyContact) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, emergencyContact); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (emergencyContact *EmergencyContact) MarshalBinary() ([]byte, error) {
	return json.Marshal(emergencyContact)
}
