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

// Campaign represents a in app campaign.
type Campaign struct {
	ID              primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt       time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt       *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt       time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy       string             `json:"createdBy" bson:"createdBy"`
	Image           string             `json:"image" bson:"image"`
	Title           string             `json:"title" bson:"title"`
	Description     string             `json:"description" bson:"description"`
	Active          bool               `json:"active" bson:"active"`
	ServiceCategory string             `json:"serviceCategory" bson:"serviceCategory"`
}

// CreateCampaign creates new campaigns.
func CreateCampaign(campaign Campaign) (*Campaign, error) {
	campaign.CreatedAt = time.Now()
	campaign.UpdatedAt = time.Now()
	campaign.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(CampaignsCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &campaign)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(campaign.ID.Hex(), campaign, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &campaign, nil
}

// GetCampaignByID gives requested campaign by id.
func GetCampaignByID(ID string) (*Campaign, error) {
	db := database.MongoDB
	campaign := &Campaign{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(campaign)
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
	err = db.Collection(CampaignsCollection).FindOne(ctx, filter).Decode(&campaign)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, campaign, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return campaign, nil
}

// GetCampaigns gives an array of campaigns.
func GetCampaigns(filter bson.D, limit int, after *string, before *string, first *int, last *int) (campaigns []*Campaign, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(CampaignsCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(CampaignsCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		campaign := &Campaign{}
		err = cur.Decode(&campaign)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		campaigns = append(campaigns, campaign)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return campaigns, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateCampaign updates the campaign.
func UpdateCampaign(c *Campaign) (*Campaign, error) {
	campaign := c
	campaign.UpdatedAt = time.Now()
	filter := bson.D{{"_id", campaign.ID}}
	db := database.MongoDB
	campaignsCollection := db.Collection(CampaignsCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := campaignsCollection.FindOneAndReplace(context.Background(), filter, campaign, findRepOpts).Decode(&campaign)
	go webhooks.NewWebhookEvent("campaign.updated", &campaign)
	if err != nil {
		log.Error(err)
	}
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(campaign.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return campaign, nil
}

// DeleteCampaignByID deletes the campaign by id.
func DeleteCampaignByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	campaignsCollection := db.Collection(CampaignsCollection)
	res, err := campaignsCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("campaign.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (campaign *Campaign) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, campaign); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (campaign *Campaign) MarshalBinary() ([]byte, error) {
	return json.Marshal(campaign)
}
