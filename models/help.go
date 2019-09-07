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

// HelpDetail represents a help detail.
type HelpDetail struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Category  HelpDetailCategory `json:"category" bson:"category"`
	Order     string             `json:"order" bson:"order"`
	Question  string             `json:"question" bson:"question"`
	Answer    string             `json:"answer" bson:"answer"`
	IsActive  bool               `json:"isActive" bson:"isActive"`
}

// CreateHelpDetail creates new help details.
func CreateHelpDetail(helpDetail HelpDetail) (*HelpDetail, error) {
	helpDetail.CreatedAt = time.Now()
	helpDetail.UpdatedAt = time.Now()
	helpDetail.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(HelpDetailsCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &helpDetail)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("help_details.created", &helpDetail)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(helpDetail.ID.Hex(), helpDetail, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &helpDetail, nil
}

// GetHelpDetailByID gives the requested help detail using id.
func GetHelpDetailByID(ID string) (*HelpDetail, error) {
	db := database.MongoDB
	helpDetail := &HelpDetail{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(helpDetail)
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
	err = db.Collection(HelpDetailsCollection).FindOne(ctx, filter).Decode(&helpDetail)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, helpDetail, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return helpDetail, nil
}

// GetHelpDetails gives an array of help details.
func GetHelpDetails(filter bson.D, limit int, after *string, before *string, first *int, last *int) (helpDetails []*HelpDetail, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(HelpDetailsCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(HelpDetailsCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		helpDetail := &HelpDetail{}
		err = cur.Decode(&helpDetail)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		helpDetails = append(helpDetails, helpDetail)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return helpDetails, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateHelpDetail updates the help detail.
func UpdateHelpDetail(c *HelpDetail) (*HelpDetail, error) {
	helpDetail := c
	helpDetail.UpdatedAt = time.Now()
	filter := bson.D{{"_id", helpDetail.ID}}
	db := database.MongoDB
	helpDetailsCollection := db.Collection(HelpDetailsCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := helpDetailsCollection.FindOneAndReplace(context.Background(), filter, helpDetail, findRepOpts).Decode(&helpDetail)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("help_details.updated", &helpDetail)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(helpDetail.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return helpDetail, nil
}

// DeleteHelpDetailByID deletes the help details by id.
func DeleteHelpDetailByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	helpDetailsCollection := db.Collection(HelpDetailsCollection)
	res, err := helpDetailsCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("help_details.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (helpDetail *HelpDetail) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, helpDetail); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (helpDetail *HelpDetail) MarshalBinary() ([]byte, error) {
	return json.Marshal(helpDetail)
}

// HelpCategory represents a help category.
type HelpCategory struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt   *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy   primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Order       string             `json:"order" bson:"order"`
	Title       string             `json:"title" bson:"title"`
	CategoryFor HelpCategoryFor    `json:"categoryFor" bson:"categoryFor"`
	IsActive    bool               `json:"isActive" bson:"isActive"`
}

// CreateHelpCategory creates new help category.
func CreateHelpCategory(helpCategory HelpCategory) (*HelpCategory, error) {
	helpCategory.CreatedAt = time.Now()
	helpCategory.UpdatedAt = time.Now()
	helpCategory.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(HelpCategoriesCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &helpCategory)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("help_category.created", &helpCategory)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(helpCategory.ID.Hex(), helpCategory, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &helpCategory, nil
}

// GetHelpCategoryByID gives requested help category by id.
func GetHelpCategoryByID(ID string) (*HelpCategory, error) {
	db := database.MongoDB
	helpCategory := &HelpCategory{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(helpCategory)
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
	err = db.Collection(HelpCategoriesCollection).FindOne(ctx, filter).Decode(&helpCategory)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, helpCategory, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return helpCategory, nil
}

// GetHelpCategories gives an array of help categories.
func GetHelpCategories(filter bson.D, limit int, after *string, before *string, first *int, last *int) (helpCategories []*HelpCategory, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(HelpCategoriesCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(HelpCategoriesCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		helpCategory := &HelpCategory{}
		err = cur.Decode(&helpCategory)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		helpCategories = append(helpCategories, helpCategory)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return helpCategories, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateHelpCategory updates the help category.
func UpdateHelpCategory(c *HelpCategory) (*HelpCategory, error) {
	helpCategory := c
	helpCategory.UpdatedAt = time.Now()
	filter := bson.D{{"_id", helpCategory.ID}}
	db := database.MongoDB
	helpCategoriesCollection := db.Collection(HelpCategoriesCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := helpCategoriesCollection.FindOneAndReplace(context.Background(), filter, helpCategory, findRepOpts).Decode(&helpCategory)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("help_category.updated", &helpCategory)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(helpCategory.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return helpCategory, nil
}

// DeleteHelpCategoryByID deletes the help categories by id.
func DeleteHelpCategoryByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	helpCategoriesCollection := db.Collection(HelpCategoriesCollection)
	res, err := helpCategoriesCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("help_category.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (helpCategory *HelpCategory) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, helpCategory); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (helpCategory *HelpCategory) MarshalBinary() ([]byte, error) {
	return json.Marshal(helpCategory)
}
