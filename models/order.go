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

// Order represents a order.
type Order struct {
	ID                           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt                    time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt                    *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt                    time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy                    primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	ParentID                     primitive.ObjectID `json:"parentID" bson:"parentID"`
	OrderNumber                  int                `json:"orderNumber" bson:"orderNumber"`
	OrderType                    string             `json:"orderType" bson:"orderType"`
	StoreID                      primitive.ObjectID `json:"storeId" bson:"storeId"`
	OrderItems                   OrderItem          `json:"orderItems" bson:"orderItems"`
	ServiceType                  string             `json:"serviceType" bson:"serviceType"`
	Coupon                       string             `json:"coupon" bson:"coupon"`
	ProviderID                   primitive.ObjectID `json:"providerId" bson:"providerId"`
	DeliveryDriver               string             `json:"deliveryDriver" bson:"deliveryDriver"`
	DeliveryAddress              Address            `json:"deliveryAddress" bson:"deliveryAddress"`
	ExpectedEarning              float64            `json:"expectedEarning" bson:"expectedEarning"`
	EarnedAmount                 float64            `json:"earnedAmount" bson:"earnedAmount"`
	CancellationAndRefundDetails string             `json:"cancellationAndRefundDetails" bson:"cancellationAndRefundDetails"`
	OrderKey                     string             `json:"orderKey" bson:"orderKey"`
	CreatedVia                   string             `json:"createdVia" bson:"createdVia"`
	Version                      string             `json:"version" bson:"version"`
	OrderStatus                  OrderStatus        `json:"orderStatus" bson:"orderStatus"`
	Currency                     Currency           `json:"currency" bson:"currency"`
	DiscountAmount               float64            `json:"discountAmount" bson:"discountAmount"`
	DiscountTax                  float64            `json:"discountTax" bson:"discountTax"`
	ShippingTotal                float64            `json:"shippingTotal" bson:"shippingTotal"`
	ShippingTax                  float64            `json:"shippingTax" bson:"shippingTax"`
	CartTax                      float64            `json:"cartTax" bson:"cartTax"`
	OrderTotalAmount             float64            `json:"orderTotalAmount" bson:"orderTotalAmount"`
	TotalTax                     float64            `json:"totalTax" bson:"totalTax"`
	PricesIncludeTax             bool               `json:"pricesIncludeTax" bson:"pricesIncludeTax"`
	CustomerID                   primitive.ObjectID `json:"customerID" bson:"customerID"`
	CustomerIPAddress            string             `json:"customerIPAddress" bson:"customerIPAddress"`
	CustomerUserAgent            string             `json:"customerUserAgent" bson:"customerUserAgent"`
	CustomerNote                 string             `json:"customerNote" bson:"customerNote"`
	Billing                      Billing            `json:"billing" bson:"billing"`
	Shipping                     Shipping           `json:"shipping" bson:"shipping"`
	PaymentMethod                PaymentMethod      `json:"paymentMethod" bson:"paymentMethod"`
	PaymentMethodTitle           string             `json:"paymentMethodTitle" bson:"paymentMethodTitle"`
	TransactionID                primitive.ObjectID `json:"transactionID" bson:"transactionID"`
	DatePaid                     time.Time          `json:"datePaid" bson:"datePaid"`
	DateCompleted                time.Time          `json:"dateCompleted" bson:"dateCompleted"`
	CartHash                     string             `json:"cartHash" bson:"cartHash"`
	Metadata                     MetaData           `json:"metadata" bson:"metadata"`
	LineItems                    LineItems          `json:"lineItems" bson:"lineItems"`
	TaxLines                     TaxLines           `json:"taxLines" bson:"taxLines"`
	ShippingLine                 ShippingLines      `json:"shippingLines" bson:"shippingLines"`
	FeeLines                     FeeLines           `json:"feeLines" bson:"feeLines"`
	CouponLines                  CouponLines        `json:"couponLines" bson:"couponLines"`
	Refunds                      Refunds            `json:"refunds" bson:"refunds"`
	IsActive                     bool               `json:"isActive" bson:"isActive"`
}

// CreateOrder creates new order.
func CreateOrder(order Order) (*Order, error) {
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()
	order.ID = primitive.NewObjectID()
	db := database.MongoDB
	ordersCollection := db.Collection(OrdersCollection)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := ordersCollection.InsertOne(ctx, &order)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("order.created", &order)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(order.ID.Hex(), order, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &order, nil
}

// GetOrderByID gives requested order by id.
func GetOrderByID(ID string) (*Order, error) {
	db := database.MongoDB
	order := &Order{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(order)
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
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	err = db.Collection(OrdersCollection).FindOne(ctx, filter).Decode(&order)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, order, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return order, nil
}

// GetOrders gives a list of orders.
func GetOrders(filter bson.D, limit int, after *string, before *string, first *int, last *int) (orders []*Order, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(OrdersCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(OrdersCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		order := &Order{}
		err = cur.Decode(&order)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		orders = append(orders, order)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return orders, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateOrder updates orders.
func UpdateOrder(o *Order) (*Order, error) {
	order := o
	order.UpdatedAt = time.Now()
	filter := bson.D{{"_id", order.ID}}
	db := database.MongoDB
	ordersCollection := db.Collection(OrdersCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := ordersCollection.FindOneAndReplace(context.Background(), filter, order, findRepOpts).Decode(&order)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("order.updated", &order)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(order.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return order, nil
}

// DeleteOrderByID deletes orders by id.
func DeleteOrderByID(ID string) (bool, error) {
	db := database.MongoDB
	filter := bson.D{{"_id", ID}}
	ordersCollection := db.Collection(OrdersCollection)
	res, err := ordersCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	log.Warningln(res)
	if err != nil && res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("order.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (order *Order) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, order); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (order *Order) MarshalBinary() ([]byte, error) {
	return json.Marshal(order)
}
