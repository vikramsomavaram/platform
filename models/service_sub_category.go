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

// ServiceSubCategory represents a service sub category.
type ServiceSubCategory struct {
	ID           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	ServiceID    string             `json:"serviceId" bson:"serviceId"` // parent service id
	CreatedAt    time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt    *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt    time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy    primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Name         string             `json:"name" bson:"name"`
	ServiceType  string             `json:"serviceType" bson:"serviceType"`
	DisplayOrder int                `json:"displayOrder" bson:"displayOrder"`
	Description  string             `json:"description" bson:"description"`
	Icon         string             `json:"icon" bson:"icon"`
	BannerImage  string             `json:"bannerImage" bson:"bannerImage"`
	IsActive     bool               `json:"isActive" bson:"isActive"`
}

// CreateServiceSubCategory creates new service sub categories.
func CreateServiceSubCategory(serviceSubCategory ServiceSubCategory) (*ServiceSubCategory, error) {
	serviceSubCategory.CreatedAt = time.Now()
	serviceSubCategory.UpdatedAt = time.Now()
	serviceSubCategory.ID = primitive.NewObjectID()
	db := database.MongoDB
	installationCollection := db.Collection(ServiceCategoriesCollection)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := installationCollection.InsertOne(ctx, &serviceSubCategory)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("service_sub_category.created", &serviceSubCategory)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(serviceSubCategory.ID.Hex(), serviceSubCategory, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &serviceSubCategory, nil
}

// GetServiceSubCategoryByID gives service sub category by id.
func GetServiceSubCategoryByID(ID string) *ServiceSubCategory {
	db := database.MongoDB
	serviceSubCategory := &ServiceSubCategory{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(serviceSubCategory)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	id, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		log.Errorln(err)
		return serviceSubCategory
	}
	filter := bson.D{{"_id", id}, {"deletedAt", bson.M{"$exists": false}}}
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	err = db.Collection(ServiceCategoriesCollection).FindOne(ctx, filter).Decode(&serviceSubCategory)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return serviceSubCategory
		}
		log.Errorln(err)
		return serviceSubCategory
	}
	//set cache item
	err = cacheClient.Set(ID, serviceSubCategory, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return serviceSubCategory
}

// GetServiceSubCategories gives a list of service sub categories.
func GetServiceSubCategories(filter bson.D, limit int, after *string, before *string, first *int, last *int) (serviceSubCategories []*ServiceSubCategory, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(ServiceCategoriesCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(ServiceCategoriesCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		serviceSubCategory := &ServiceSubCategory{}
		err = cur.Decode(&serviceSubCategory)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		serviceSubCategories = append(serviceSubCategories, serviceSubCategory)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return serviceSubCategories, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateServiceSubCategory updates survice sub category.
func UpdateServiceSubCategory(s *ServiceSubCategory) (*ServiceSubCategory, error) {
	serviceSubCategory := s
	serviceSubCategory.UpdatedAt = time.Now()
	filter := bson.D{{"_id", serviceSubCategory.ID}}
	db := database.MongoDB
	serviceCategoriesCollection := db.Collection(ServiceCategoriesCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := serviceCategoriesCollection.FindOneAndReplace(context.Background(), filter, serviceSubCategory, findRepOpts).Decode(&serviceSubCategory)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("service_sub_category.updated", &serviceSubCategory)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(serviceSubCategory.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return serviceSubCategory, nil
}

// DeleteServiceSubCategoryByID deletes service sub category.
func DeleteServiceSubCategoryByID(ID string) (bool, error) {
	db := database.MongoDB
	id, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", id}}
	serviceCategoriesCollection := db.Collection(ServiceCategoriesCollection)
	res, err := serviceCategoriesCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil && res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("service_sub_category.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (serviceSubCategory *ServiceSubCategory) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, serviceSubCategory); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (serviceSubCategory *ServiceSubCategory) MarshalBinary() ([]byte, error) {
	return json.Marshal(serviceSubCategory)
}
