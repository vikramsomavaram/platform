/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package models

import (
	"context"
	"crypto/sha1"
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

type Cart struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	UserID    primitive.ObjectID `json:"userID" bson:"userID"`
	StoreID   string             `json:"storeID" bson:"storeID"`
	Items     []CartItem         `json:"cartItems" bson:"cartItems"`
}

type CartItem struct {
	ProductID   primitive.ObjectID  `json:"productID" bson:"productID"`
	Quantity    int                 `json:"quantity" bson:"quantity"`
	VariationID *primitive.ObjectID `json:"variationID" bson:"variationID"`
	Type        CartItemType        `json:"type" bson:"type"`
}

// CreateCart creates new cart.
func CreateCart(cart *Cart) (*Cart, error) {
	cart.CreatedAt = time.Now()
	cart.UpdatedAt = time.Now()
	cart.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(CartCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &cart)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("cart.created", &cart)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(cart.ID.Hex(), cart, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return cart, nil
}

// GetCartByID gets carts by ID.
func GetCartByID(ID string) *Cart {
	db := database.MongoDB
	cart := &Cart{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(cart)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}

	oID, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		log.Errorln(err)
		return cart
	}
	filter := bson.D{{"_id", oID}, {"deletedAt", bson.M{"$exists": false}}}
	ctx := context.Background()
	err = db.Collection(CartCollection).FindOne(ctx, filter).Decode(&cart)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return cart
		}
		return cart
	}
	//set cache item
	err = cacheClient.Set(ID, cart, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return cart
}

// GetCartByID gets carts by ID.
func GetCartByFilter(filter bson.D) (*Cart, error) {
	db := database.MongoDB
	cart := &Cart{}
	filter = append(filter, bson.E{"deletedAt", bson.M{"$exists": false}})
	err, filterHash := genBsonHash(filter)
	if err != nil {
		return nil, err
	}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err = cacheClient.Get(filterHash).Scan(cart)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}

	ctx := context.Background()
	err = db.Collection(CartCollection).FindOne(ctx, filter).Decode(&cart)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		log.Error(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(filterHash, cart, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	err = cacheClient.Set(cart.ID.Hex(), cart, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return cart, nil
}

func genBsonHash(filter bson.D) (error, string) {
	b, err := bson.Marshal(filter)
	if err != nil {
		log.Errorln(err)
		return err, ""
	}
	h := sha1.New()
	h.Write(b)
	bs := h.Sum(nil)
	return nil, string(bs)
}

// GetCarts gets the array of carts.
func GetCarts(filter bson.D, limit int, after *string, before *string, first *int, last *int) (carts []*Cart, totalCount int64, hasPrevious, hasNext bool, err error) {
	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(CartCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(CartCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		cart := &Cart{}
		err = cur.Decode(&cart)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		carts = append(carts, cart)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return carts, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateCart updates the carts.
func UpdateCart(c *Cart) (*Cart, error) {
	cart := c
	cart.UpdatedAt = time.Now()
	filter := bson.D{{"_id", cart.ID}}
	db := database.MongoDB
	collection := db.Collection(CartCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := collection.FindOneAndReplace(context.Background(), filter, cart, findRepOpts).Decode(&cart)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("cart.updated", &cart)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(cart.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return cart, nil
}

// DeleteCartByID deletes the carts by ID.
func DeleteCartByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	collection := db.Collection(CartCollection)
	res, err := collection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	log.Warningln(res)
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("cart.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (cart *Cart) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, cart); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (cart *Cart) MarshalBinary() ([]byte, error) {
	return json.Marshal(cart)
}
