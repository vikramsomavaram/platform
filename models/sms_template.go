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

// SMSTemplate represents a sms template.
type SMSTemplate struct {
	ID         primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt  time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt  time.Time          `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt  time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy  primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Title      string             `json:"title" bson:"title"`
	Body       string             `json:"body" bson:"body"`
	Code       string             `json:"code" bson:"code"`
	Language   string             `json:"language" bson:"language"`
	Purpose    string             `json:"purpose" bson:"purpose" `
	TemplateID string             `json:"templateId" bson:"templateId"`
}

// CreateSMSTemplate creates new sms template.
func CreateSMSTemplate(smsTemplate *SMSTemplate) (*SMSTemplate, error) {
	smsTemplate.CreatedAt = time.Now()
	smsTemplate.UpdatedAt = time.Now()
	smsTemplate.ID = primitive.NewObjectID()
	db := database.MongoDB
	smsTemplateCollection := db.Collection(SMSTemplateCollection)
	ctx := context.Background()
	_, err := smsTemplateCollection.InsertOne(ctx, &smsTemplate)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(smsTemplate.ID.Hex(), smsTemplate, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("sms_template.created", &smsTemplate)
	return smsTemplate, nil
}

// GetSMSTemplateByID gives sms template by id.
func GetSMSTemplateByID(ID string) (*SMSTemplate, error) {
	db := database.MongoDB
	smsTemplate := &SMSTemplate{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(smsTemplate)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}

	if !smsTemplate.ID.IsZero() {
		return smsTemplate, nil
	}
	oID, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		return nil, err
	}
	filter := bson.D{{"_id", oID}, {"deletedAt", bson.M{"$exists": false}}}
	ctx := context.Background()
	err = db.Collection(SMSTemplateCollection).FindOne(ctx, filter).Decode(&smsTemplate)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, smsTemplate, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return smsTemplate, nil
}

// GetSMSTemplates gives a list of sms templates.
func GetSMSTemplates(filter bson.D, limit int, after *string, before *string, first *int, last *int) (smsTemplates []*SMSTemplate, totalCount int64, hasPrevious, hasNext bool, err error) {
	db := database.MongoDB
	tcint, filter, err := calcTotalCountWithQueryFilters(SMSTemplateCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(SMSTemplateCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
ctx := context.Background()
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		smsTemplate := &SMSTemplate{}
		err = cur.Decode(&smsTemplate)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		smsTemplates = append(smsTemplates, smsTemplate)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return smsTemplates, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateSMSTemplate updates sms template.
func UpdateSMSTemplate(e *SMSTemplate) (*SMSTemplate, error) {
	smsTemplate := e
	smsTemplate.UpdatedAt = time.Now()
	filter := bson.D{{"_id", smsTemplate.ID}}
	db := database.MongoDB
	smsTemplateCollection := db.Collection(SMSTemplateCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := smsTemplateCollection.FindOneAndReplace(context.Background(), filter, smsTemplate, findRepOpts).Decode(&smsTemplate)
	if err != nil {
		log.Error(err)
	}
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(smsTemplate.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("sms_template.updated", &smsTemplate)
	return smsTemplate, nil
}

// DeleteSMSTemplateByID deletes sms template by id.
func DeleteSMSTemplateByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	smsTemplateCollection := db.Collection(SMSTemplateCollection)
	res, err := smsTemplateCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
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
	go webhooks.NewWebhookEvent("sms_template.deleted", &res)
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (smsTemplate *SMSTemplate) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, smsTemplate); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (smsTemplate *SMSTemplate) MarshalBinary() ([]byte, error) {
	return json.Marshal(smsTemplate)
}
