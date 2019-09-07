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

type OrderNote struct {
	ID           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt    time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt    *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt    time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy    primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Author       string             `json:"author" bson:"author"`
	Note         string             `json:"note" bson:"note"`
	CustomerNote bool               `json:"customerNote" bson:"customerNote"`
	IsActive     bool               `json:"isActive" bson:"isActive"`
}

// CreateOrderNote creates new order note.
func CreateOrderNote(orderNote OrderNote) (*OrderNote, error) {
	orderNote.CreatedAt = time.Now()
	orderNote.UpdatedAt = time.Now()
	orderNote.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(OrderNotesCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &orderNote)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("order_note.created", &orderNote)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(orderNote.ID.Hex(), orderNote, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &orderNote, nil
}

// GetOrderNoteByID gives order note by id.
func GetOrderNoteByID(ID string) (*OrderNote, error) {
	db := database.MongoDB
	orderNote := &OrderNote{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(orderNote)
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
	err = db.Collection(OrderNotesCollection).FindOne(ctx, filter).Decode(&orderNote)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, orderNote, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return orderNote, nil
}

// GetOrderNotes gives a list of order notes.
func GetOrderNotes(filter bson.D, limit int, after *string, before *string, first *int, last *int) (orderNotes []*OrderNote, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(OrderNotesCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(OrderNotesCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		orderNote := &OrderNote{}
		err = cur.Decode(&orderNote)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		orderNotes = append(orderNotes, orderNote)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return orderNotes, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateOrderNote updates order note.
func UpdateOrderNote(s *OrderNote) (*OrderNote, error) {
	orderNote := s
	orderNote.UpdatedAt = time.Now()
	filter := bson.D{{"_id", orderNote.ID}}
	db := database.MongoDB
	orderNotesCollection := db.Collection(OrderNotesCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := orderNotesCollection.FindOneAndReplace(context.Background(), filter, orderNote, findRepOpts).Decode(&orderNote)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("order_note.updated", &orderNote)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(orderNote.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return orderNote, nil
}

// DeleteOrderNoteByID deletes order note by id.
func DeleteOrderNoteByID(ID string) (bool, error) {
	db := database.MongoDB
	id, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", id}}
	orderNotesCollection := db.Collection(OrderNotesCollection)
	res, err := orderNotesCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil && res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("order_note.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (orderNote *OrderNote) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, orderNote); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (orderNote *OrderNote) MarshalBinary() ([]byte, error) {
	return json.Marshal(orderNote)
}
