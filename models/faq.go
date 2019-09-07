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

// FAQ represents an faq.
type FAQ struct {
	ID           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt    time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt    *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt    time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy    primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Category     primitive.ObjectID `json:"category" bson:"category"`
	DisplayOrder int                `json:"displayOrder" bson:"displayOrder"`
	Question     string             `json:"question" bson:"question"`
	Answer       string             `json:"answer" bson:"answer"`
	IsActive     bool               `json:"isActive" bson:"isActive"`
}

// CreateFAQ creates new faqs.
func CreateFAQ(faq FAQ) (*FAQ, error) {
	faq.CreatedAt = time.Now()
	faq.UpdatedAt = time.Now()
	faq.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(FAQsCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &faq)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("faq.created", &faq)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(faq.ID.Hex(), faq, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &faq, nil
}

// GetFAQByID gives the requested faq using id.
func GetFAQByID(ID string) (*FAQ, error) {
	db := database.MongoDB
	faq := &FAQ{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(faq)
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
	err = db.Collection(FAQsCollection).FindOne(ctx, filter).Decode(&faq)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, faq, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return faq, nil
}

// GetFAQs gives an array of faqs.
func GetFAQs(filter bson.D, limit int, after *string, before *string, first *int, last *int) (faqs []*FAQ, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(FAQsCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(FAQsCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
ctx := context.Background()
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		faq := &FAQ{}
		err = cur.Decode(&faq)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		faqs = append(faqs, faq)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return faqs, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateFAQ updates faq.
func UpdateFAQ(c *FAQ) (*FAQ, error) {
	faq := c
	faq.UpdatedAt = time.Now()
	filter := bson.D{{"_id", faq.ID}}
	db := database.MongoDB
	faqsCollection := db.Collection(FAQsCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := faqsCollection.FindOneAndReplace(context.Background(), filter, faq, findRepOpts).Decode(&faq)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("faq.updated", &faq)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(faq.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return faq, nil
}

// DeleteFAQByID deletes the faq by id.
func DeleteFAQByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	faqsCollection := db.Collection(FAQsCollection)
	res, err := faqsCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("faq.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (faq *FAQ) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, faq); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (faq *FAQ) MarshalBinary() ([]byte, error) {
	return json.Marshal(faq)
}

// FAQCategory represents a faq category.
type FAQCategory struct {
	ID           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt    time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt    *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt    time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy    primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	DisplayOrder int                `json:"displayOrder" bson:"displayOrder"`
	Label        string             `json:"label" bson:"label"`
	IsActive     bool               `json:"isActive" bson:"isActive"`
}

// CreateFAQCategory creates new faq categories.
func CreateFAQCategory(faqCategory FAQCategory) (*FAQCategory, error) {
	faqCategory.CreatedAt = time.Now()
	faqCategory.UpdatedAt = time.Now()
	faqCategory.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(FAQsCategoryCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &faqCategory)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("faq_category.created", &faqCategory)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(faqCategory.ID.Hex(), faqCategory, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &faqCategory, nil
}

// GetFAQCategoryByID gives the requested faq using id.
func GetFAQCategoryByID(ID string) (*FAQCategory, error) {
	db := database.MongoDB
	faqCategory := &FAQCategory{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(faqCategory)
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
	err = db.Collection(FAQsCategoryCollection).FindOne(ctx, filter).Decode(&faqCategory)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, faqCategory, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return faqCategory, nil
}

// GetFAQCategories gives an array of faq categories.
func GetFAQCategories(filter bson.D, limit int, after *string, before *string, first *int, last *int) (faqCategories []*FAQCategory, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(FAQsCategoryCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(FAQsCategoryCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
ctx := context.Background()
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		faqCategory := &FAQCategory{}
		err = cur.Decode(&faqCategory)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		faqCategories = append(faqCategories, faqCategory)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return faqCategories, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateFAQCategory updates the faq cateories.
func UpdateFAQCategory(c *FAQCategory) (*FAQCategory, error) {
	faqCategory := c
	faqCategory.UpdatedAt = time.Now()
	filter := bson.D{{"_id", faqCategory.ID}}
	db := database.MongoDB
	faqCategoriesCollection := db.Collection(FAQsCategoryCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := faqCategoriesCollection.FindOneAndReplace(context.Background(), filter, faqCategory, findRepOpts).Decode(&faqCategory)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("faq_category.updated", &faqCategory)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(faqCategory.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return faqCategory, nil
}

// DeleteFAQCategoryByID deletes the faq category by id.
func DeleteFAQCategoryByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	faqCategoriesCollection := db.Collection(FAQsCategoryCollection)
	res, err := faqCategoriesCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil && res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("faq_category.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (faqCategory *FAQCategory) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, faqCategory); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (faqCategory *FAQCategory) MarshalBinary() ([]byte, error) {
	return json.Marshal(faqCategory)
}
