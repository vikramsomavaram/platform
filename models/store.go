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

// Store represents a store.
type Store struct {
	ID                       primitive.ObjectID  `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt                time.Time           `json:"createdAt" bson:"createdAt"`
	DeletedAt                *time.Time          `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt                time.Time           `json:"updatedAt" bson:"updatedAt"`
	CreatedBy                primitive.ObjectID  `json:"createdBy" bson:"createdBy"`
	StoreName                string              `json:"storeName" bson:"storeName"`
	StoreLocation            StoreLocation       `json:"storeLocation" bson:"storeLocation"`
	ServiceCategory          StoreCategory       `json:"serviceCategory" bson:"serviceCategory"`
	Email                    string              `json:"email" bson:"email"`
	Password                 string              `json:"password" bson:"password"`
	StoreAddress             Address             `json:"storeAddress" bson:"storeAddress"`
	ZipCode                  string              `json:"zipCode" bson:"zipCode"`
	Country                  string              `json:"country" bson:"country"`
	State                    string              `json:"state" bson:"state"`
	ContactPersonName        string              `json:"contactPersonName" bson:"contactPersonName"`
	MobileNumber             string              `json:"mobileNumber" bson:"mobileNumber"`
	StoreLogo                string              `json:"storeLogo" bson:"storeLogo"`
	Language                 string              `json:"language" bson:"language"`
	AvailableStoreItemTypes  string              `json:"availableStoreItemTypes" bson:"availableStoreItemTypes"`
	Slot1                    time.Time           `json:"slot1" bson:"slot1"`
	Slot2                    time.Time           `json:"slot2" bson:"slot2"`
	MinimumAmountPerOrder    float64             `json:"minimumAmountPerOrder" bson:"minimumAmountPerOrder"`
	AdditionalPackingCharges float64             `json:"additionalPackingCharges" bson:"additionalPackingCharges"`
	MaxOrderQuantity         string              `json:"maxOrderQuantity" bson:"maxOrderQuantity"`
	EstimatedOrderTime       int                 `json:"estimatedOrderTime" bson:"estimatedOrderTime"`
	OfferAppliesOn           OfferAppliesOn      `json:"offerAppliesOn" bson:"offerAppliesOn"`
	BankAccountDetails       primitive.ObjectID  `json:"bankAccountDetails" bson:"bankAccountDetails"`
	IsActive                 bool                `json:"isActive" bson:"isActive"`
	IsMultiLocationEnabled   bool                `json:"isMultiLocationEnabled"`
	Blocked                  bool                `json:"blocked" bson:"blocked"`
	ApprovedAt               *time.Time          `json:"approvedAt" bson:"approvedAt"`
	ApprovedBy               *primitive.ObjectID `json:"approvedBy" bson:"approvedBy"`
}

type StoreLocation struct {
	ID                primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt         time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt         *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt         time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy         primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	StoreID           primitive.ObjectID `json:"storeID" bson:"storeID"`
	StoreLocationName string             `json:"storeLocation" bson:"storeLocation"`
	StoreAddress      Address            `json:"storeAddress" bson:"storeAddress"`
}

// CreateStore creates a store.
func CreateStore(store Store) (*Store, error) {
	store.CreatedAt = time.Now()
	store.UpdatedAt = time.Now()
	store.ID = primitive.NewObjectID()
	db := database.MongoDB
	installationCollection := db.Collection(StoresCollection)
	ctx := context.Background()
	_, err := installationCollection.InsertOne(ctx, &store)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("store.created", &store)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(store.ID.Hex(), store, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &store, nil
}

// GetStoreByID gives a store by id.
func GetStoreByID(ID string) *Store {
	db := database.MongoDB
	store := &Store{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(store)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	id, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		log.Errorln(err)
		return store
	}
	filter := bson.D{{"_id", id}, {"deletedAt", bson.M{"$exists": false}}}
	ctx := context.Background()
	err = db.Collection(StoresCollection).FindOne(ctx, filter).Decode(&store)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return store
		}
		log.Errorln(err)
		return store
	}
	//set cache item
	err = cacheClient.Set(ID, store, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return store
}

// GetStores gives a list of stores.
func GetStores(filter bson.D, limit int, after *string, before *string, first *int, last *int) (stores []*Store, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(StoresCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(StoresCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		store := &Store{}
		err = cur.Decode(&store)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		stores = append(stores, store)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return stores, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateStore updates store.
func UpdateStore(s *Store) (*Store, error) {
	store := s
	store.UpdatedAt = time.Now()
	filter := bson.D{{"_id", store.ID}}
	db := database.MongoDB
	storesCollection := db.Collection(StoresCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := storesCollection.FindOneAndReplace(context.Background(), filter, store, findRepOpts).Decode(&store)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("store.updated", &store)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(store.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return store, nil
}

// DeleteStoreByID deletes store by id.
func DeleteStoreByID(ID string) (bool, error) {
	db := database.MongoDB
	id, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", id}}
	storesCollection := db.Collection(StoresCollection)
	res, err := storesCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil && res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("store.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

// GetStoreByFilter gives a store by query filter.
func GetStoreByFilter(filter bson.D) (*Store, error) {
	db := database.MongoDB
	store := &Store{}
	err, filterHash := genBsonHash(filter)
	if err != nil {
		return nil, err
	}

	//try finding item in cache
	cacheClient := cache.RedisClient
	err = cacheClient.Get(filterHash).Scan(store)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}

	filter = append(filter, bson.E{"deletedAt", bson.M{"$exists": false}})
	ctx := context.Background()
	err = db.Collection(StoresCollection).FindOne(ctx, filter).Decode(&store)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return store, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(filterHash, store, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return store, nil
}

//UnmarshalBinary required for the redis cache to work
func (store *Store) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, store); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (store *Store) MarshalBinary() ([]byte, error) {
	return json.Marshal(store)
}

// CreateStoreLocation creates a store location.
func CreateStoreLocation(store StoreLocation) (*StoreLocation, error) {
	store.CreatedAt = time.Now()
	store.UpdatedAt = time.Now()
	store.ID = primitive.NewObjectID()
	db := database.MongoDB
	installationCollection := db.Collection(StoreLocationsCollection)
	ctx := context.Background()
	_, err := installationCollection.InsertOne(ctx, &store)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("storeLocation.created", &store)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(store.ID.Hex(), store, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &store, nil
}

// GetStoreLocationByID gives a store location by id.
func GetStoreLocationByID(ID string) (*StoreLocation, error) {
	db := database.MongoDB
	store := &StoreLocation{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(store)
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
	err = db.Collection(StoreLocationsCollection).FindOne(ctx, filter).Decode(&store)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, store, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return store, nil
}

// GetStoreLocations gives a list of store locations.
func GetStoreLocations(filter bson.D, limit int, after *string, before *string, first *int, last *int) (stores []*StoreLocation, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(StoreLocationsCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(StoreLocationsCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		store := &StoreLocation{}
		err = cur.Decode(&store)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		stores = append(stores, store)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return stores, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateStoreLocation updates store location.
func UpdateStoreLocation(store *StoreLocation) (*StoreLocation, error) {
	store.UpdatedAt = time.Now()
	filter := bson.D{{"_id", store.ID}}
	db := database.MongoDB
	storesCollection := db.Collection(StoreLocationsCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := storesCollection.FindOneAndReplace(context.Background(), filter, store, findRepOpts).Decode(&store)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("storeLocation.updated", &store)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(store.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return store, nil
}

// DeleteStoreLocationByID deletes store location by id.
func DeleteStoreLocationByID(ID string) (bool, error) {
	db := database.MongoDB
	id, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", id}}
	storesCollection := db.Collection(StoreLocationsCollection)
	res, err := storesCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil && res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("storeLocation.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}
