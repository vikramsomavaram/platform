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

// Coupon represents a coupon.
type Coupon struct {
	ID                        primitive.ObjectID   `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt                 time.Time            `json:"createdAt" bson:"createdAt"`
	DeletedAt                 time.Time            `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt                 time.Time            `json:"updatedAt" bson:"updatedAt"`
	CreatedBy                 primitive.ObjectID   `json:"createdBy" bson:"createdBy"`
	Code                      string               `json:"code" bson:"code"`
	Description               string               `json:"description" bson:"description"`
	DiscountAmount            float64              `json:"discountAmount" bson:"discountAmount"`
	DiscountType              string               `json:"discountType" bson:"discountType"` //percentage / flat
	Validity                  string               `json:"validity" bson:"validity"`         //permanent or set timings
	ValidityStart             time.Time            `json:"validityStart" bson:"validityStart"`
	ValidityExpire            time.Time            `json:"validityExpire" bson:"validityExpire"`
	UsageLimit                int                  `json:"usageLimit" bson:"usageLimit"`
	UsedLimit                 int                  `json:"usedLimit" bson:"usedLimit"`
	Type                      CouponType           `json:"type" bson:"type"`
	ServiceType               CouponSystemType     `json:"serviceType" bson:"serviceType"`
	IsActive                  bool                 `json:"isActive" bson:"isActive"`
	IndividualUse             bool                 `json:"individualUse"`
	ProductIds                []primitive.ObjectID `json:"productIDs"`
	ExcludedProductIds        []primitive.ObjectID `json:"excludedProductIDs"`
	UsageLimitPerUser         int                  `json:"usageLimitPerUser"`
	LimitUsageToXItems        int                  `json:"limitUsageToXItems"`
	FreeShipping              bool                 `json:"freeShipping"`
	ProductCategories         []primitive.ObjectID `json:"productCategories"`
	ExcludedProductCategories []primitive.ObjectID `json:"excludedProductCategories"`
	ExcludeSaleItems          bool                 `json:"excludeSaleItems"`
	MinimumAmount             float64              `json:"minimumAmount"`
	MaximumAmount             float64              `json:"maximumAmount"`
	EmailRestrictions         []string             `json:"emailRestrictions"`
	UsedBy                    []primitive.ObjectID `json:"usedBy"`
	MetaData                  MetaData             `json:"metaData"`
}

// CreateCoupon creates new coupon.
func CreateCoupon(coupon Coupon) (*Coupon, error) {
	coupon.CreatedAt = time.Now()
	coupon.UpdatedAt = time.Now()
	coupon.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(CouponCollection)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := collection.InsertOne(ctx, &coupon)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("coupon.created", &coupon)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(coupon.ID.Hex(), coupon, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &coupon, nil
}

// GetCouponByFilter gives requested coupon by filter.
func GetCouponByFilter(filter bson.D) *Coupon {
	db := database.MongoDB
	coupon := &Coupon{}
	err, filterHash := genBsonHash(filter)
	if err != nil {
		log.Errorln(err)
	}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err = cacheClient.Get(filterHash).Scan(coupon)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	filter = append(filter, bson.E{"deletedAt", bson.M{"$exists": false}})
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	err = db.Collection(CouponCollection).FindOne(ctx, filter).Decode(&coupon)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return coupon
		}
		log.Errorln(err)
		return coupon
	}
	//set cache item
	err = cacheClient.Set(filterHash, coupon, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return coupon
}

// GetCouponByID gives requested coupon by id.
func GetCouponByID(ID string) *Coupon {
	db := database.MongoDB
	coupon := &Coupon{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(coupon)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}

	id, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		log.Errorln(err)
		return coupon
	}
	filter := bson.D{{"_id", id}, {"deletedAt", bson.M{"$exists": false}}}
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	err = db.Collection(CouponCollection).FindOne(ctx, filter).Decode(&coupon)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return coupon
		}
		log.Errorln(err)
		return coupon
	}
	//set cache item
	err = cacheClient.Set(ID, coupon, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return coupon
}

// GetCoupons gives a list of coupons.
func GetCoupons(filter bson.D, limit int, after *string, before *string, first *int, last *int) (coupons []*Coupon, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(CouponCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(CouponCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		coupon := &Coupon{}
		err = cur.Decode(&coupon)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		coupons = append(coupons, coupon)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return coupons, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateCoupon updates coupon.
func UpdateCoupon(c *Coupon) (*Coupon, error) {
	coupon := c
	coupon.UpdatedAt = time.Now()
	filter := bson.D{{"_id", coupon.ID}}
	db := database.MongoDB
	couponCollection := db.Collection(CouponCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := couponCollection.FindOneAndReplace(context.Background(), filter, coupon, findRepOpts).Decode(&coupon)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("coupon.updated", &coupon)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(coupon.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return coupon, nil
}

// DeleteCouponByID deletes coupon by id.
func DeleteCouponByID(ID string) (bool, error) {
	db := database.MongoDB
	id, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", id}}
	couponCollection := db.Collection(CouponCollection)
	res, err := couponCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("coupon.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (coupon *Coupon) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, coupon); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (coupon *Coupon) MarshalBinary() ([]byte, error) {
	return json.Marshal(coupon)
}
