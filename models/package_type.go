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

// PackageType represents a package type.
type PackageType struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt   *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy   primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	PackageType string             `json:"packageType" bson:"packageType"`
	Language    string             `json:"language" bson:"language"`
	IsActive    bool               `json:"isActive" bson:"isActive"`
}

// CreatePackageType creates new package type.
func CreatePackageType(packageType PackageType) (*PackageType, error) {
	packageType.CreatedAt = time.Now()
	packageType.UpdatedAt = time.Now()
	packageType.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(PackageTypeCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &packageType)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("package_type.created", &packageType)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(packageType.ID.Hex(), packageType, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &packageType, nil
}

// GetPackageTypeByID gives requested package type by id.
func GetPackageTypeByID(ID string) (*PackageType, error) {
	db := database.MongoDB
	packageType := &PackageType{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(packageType)
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
	err = db.Collection(PackageTypeCollection).FindOne(ctx, filter).Decode(&packageType)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, packageType, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return packageType, nil
}

// GetPackageTypes gives a list of package types.
func GetPackageTypes(filter bson.D, limit int, after *string, before *string, first *int, last *int) (packageTypes []*PackageType, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(PackageTypeCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(PackageTypeCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
ctx := context.Background()
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		packageType := &PackageType{}
		err = cur.Decode(&packageType)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		packageTypes = append(packageTypes, packageType)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return packageTypes, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdatePackageType updates package type.
func UpdatePackageType(p *PackageType) (*PackageType, error) {
	packageType := p
	packageType.UpdatedAt = time.Now()
	filter := bson.D{{"_id", packageType.ID}}
	db := database.MongoDB
	packageTypeCollection := db.Collection(PackageTypeCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := packageTypeCollection.FindOneAndReplace(context.Background(), filter, packageType, findRepOpts).Decode(&packageType)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("package_type.updated", &packageType)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(packageType.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return packageType, nil
}

// DeletePackageTypeByID deletes package type by id.
func DeletePackageTypeByID(ID string) (bool, error) {
	db := database.MongoDB
	id, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", id}}
	packageTypeCollection := db.Collection(PackageTypeCollection)
	res, err := packageTypeCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("package_type.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (packageType *PackageType) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, packageType); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (packageType *PackageType) MarshalBinary() ([]byte, error) {
	return json.Marshal(packageType)
}
