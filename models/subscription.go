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

// PlanInterval is the list of allowed values for a plan's interval.
type SubscriptionPlanInterval string

// List of values that PlanInterval can take.
const (
	SubscriptionPlanIntervalDay   SubscriptionPlanInterval = "day"
	SubscriptionPlanIntervalWeek  SubscriptionPlanInterval = "week"
	SubscriptionPlanIntervalMonth SubscriptionPlanInterval = "month"
	SubscriptionPlanIntervalYear  SubscriptionPlanInterval = "year"
)

// PlanBillingScheme is the list of allowed values for a plan's billing scheme.
type SubscriptionPlanBillingScheme string

// List of values that PlanBillingScheme can take.
const (
	SubscriptionPlanBillingSchemePerUnit SubscriptionPlanBillingScheme = "perUnit"
	SubscriptionPlanBillingSchemeTiered  SubscriptionPlanBillingScheme = "tiered"
)

// PlanUsageType is the list of allowed values for a plan's usage type.
type SubscriptionPlanUsageType string

// List of values that PlanUsageType can take.
const (
	SubscriptionPlanUsageTypeLicensed SubscriptionPlanUsageType = "licensed"
	SubscriptionPlanUsageTypeMetered  SubscriptionPlanUsageType = "metered"
)

// PlanTiersMode is the list of allowed values for a plan's tiers mode.
type SubscriptionPlanTiersMode string

// List of values that PlanTiersMode can take.
const (
	SubscriptionPlanTiersModeGraduated SubscriptionPlanTiersMode = "graduated"
	SubscriptionPlanTiersModeVolume    SubscriptionPlanTiersMode = "volume"
)

// PlanTransformUsageRound is the list of allowed values for a plan's transform usage round logic.
type SubscriptionPlanTransformUsageRound string

// List of values that PlanTransformUsageRound can take.
const (
	SubscriptionPlanTransformUsageRoundDown SubscriptionPlanTransformUsageRound = "down"
	SubscriptionPlanTransformUsageRoundUp   SubscriptionPlanTransformUsageRound = "up"
)

// PlanAggregateUsage is the list of allowed values for a plan's aggregate usage.
type SubscriptionPlanAggregateUsage string

// List of values that PlanAggregateUsage can take.
const (
	SubscriptionPlanAggregateUsageLastDuringPeriod SubscriptionPlanAggregateUsage = "lastDuringPeriod"
	SubscriptionPlanAggregateUsageLastEver         SubscriptionPlanAggregateUsage = "lastEver"
	SubscriptionPlanAggregateUsageMax              SubscriptionPlanAggregateUsage = "max"
	SubscriptionPlanAggregateUsageSum              SubscriptionPlanAggregateUsage = "sum"
)

// PlanTier configures tiered pricing
type SubscriptionPlanTier struct {
	FlatAmount int64 `json:"flat_amount"`
	UnitAmount int64 `json:"unit_amount"`
	UpTo       int64 `json:"up_to"`
}

type SubscriptionPlan struct {
	ID              primitive.ObjectID            `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt       time.Time                     `json:"createdAt" bson:"createdAt"`
	DeletedAt       *time.Time                    `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt       time.Time                     `json:"updatedAt" bson:"updatedAt"`
	CreatedBy       primitive.ObjectID            `json:"createdBy" bson:"createdBy"`
	Active          bool                          `json:"active"`
	AggregateUsage  string                        `json:"aggregateUsage"`
	Amount          int64                         `json:"amount"`
	BillingScheme   SubscriptionPlanBillingScheme `json:"billingScheme"`
	Created         int64                         `json:"created"`
	Currency        Currency                      `json:"currency"`
	Deleted         bool                          `json:"deleted"`
	Interval        SubscriptionPlanInterval      `json:"interval"`
	IntervalCount   int64                         `json:"intervalCount"`
	Livemode        bool                          `json:"livemode"`
	Metadata        map[string]string             `json:"metadata"`
	Nickname        string                        `json:"nickname"`
	Product         *Product                      `json:"product"`
	Tiers           []*SubscriptionPlanTier       `json:"tiers"`
	TiersMode       string                        `json:"tiersMode"`
	TransformUsage  *PlanTransformUsage           `json:"transformUsage"`
	TrialPeriodDays int64                         `json:"trialPeriodDays"`
	UsageType       SubscriptionPlanUsageType     `json:"usageType"`
}

// PlanTransformUsage represents the bucket billing configuration.
type PlanTransformUsage struct {
	DivideBy int64                               `json:"divideBy"`
	Round    SubscriptionPlanTransformUsageRound `json:"round"`
}

// SubscriptionBillingThresholds is a structure representing the billing thresholds for a subscription.
type SubscriptionBillingThresholds struct {
	AmountGTE               int64 `json:"amountGte"`
	ResetBillingCycleAnchor bool  `json:"resetBillingCycleAnchor"`
}

// SubscriptionStatus is the list of allowed values for the subscription's status.
type SubscriptionStatus string

// List of values that SubscriptionStatus can take.
const (
	SubscriptionStatusActive            SubscriptionStatus = "active"
	SubscriptionStatusAll               SubscriptionStatus = "all"
	SubscriptionStatusCanceled          SubscriptionStatus = "canceled"
	SubscriptionStatusIncomplete        SubscriptionStatus = "incomplete"
	SubscriptionStatusIncompleteExpired SubscriptionStatus = "incomplete_expired"
	SubscriptionStatusPastDue           SubscriptionStatus = "past_due"
	SubscriptionStatusTrialing          SubscriptionStatus = "trialing"
	SubscriptionStatusUnpaid            SubscriptionStatus = "unpaid"
)

// SubscriptionBilling is the type of billing method for this subscription's invoices.
type SubscriptionBilling string

// List of values that SubscriptionBilling can take.
const (
	SubscriptionBillingChargeAutomatically SubscriptionBilling = "charge_automatically"
	SubscriptionBillingSendInvoice         SubscriptionBilling = "send_invoice"
)

// TaxRate is the resource representing a Stripe tax rate.
// For more details see https://stripe.com/docs/api/tax_rates/object.
type TaxRate struct {
	Active       bool               `json:"active"`
	Description  string             `json:"description"`
	DisplayName  string             `json:"display_name"`
	ID           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt    time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt    *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt    time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy    primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Inclusive    bool               `json:"inclusive"`
	Jurisdiction string             `json:"jurisdiction"`
	Livemode     bool               `json:"livemode"`
	Metadata     map[string]string  `json:"metadata"`
	Percentage   float64            `json:"percentage"`
}

type Discount struct {
	Coupon       *Coupon   `json:"coupon"`
	Customer     string    `json:"customer"`
	Deleted      bool      `json:"deleted"`
	End          time.Time `json:"end"`
	Start        time.Time `json:"start"`
	Subscription string    `json:"subscription"`
}

// PaymentSourceType consts represent valid payment sources.
type PaymentSourceType string

// List of values that PaymentSourceType can take.
const (
	PaymentSourceTypeAccount         PaymentSourceType = "account"
	PaymentSourceTypeBankAccount     PaymentSourceType = "bank_account"
	PaymentSourceTypeBitcoinReceiver PaymentSourceType = "bitcoin_receiver"
	PaymentSourceTypeCard            PaymentSourceType = "card"
	PaymentSourceTypeObject          PaymentSourceType = "source"
)

// PaymentSource describes the payment source used to make a Charge.
// The Type should indicate which object is fleshed out (eg. BitcoinReceiver or Card)
// For more details see https://stripe.com/docs/api#retrieve_charge
type PaymentSource struct {
	BankAccount *BankAccount `json:"-"`
	Card        *Card        `json:"-"`
	Deleted     bool         `json:"deleted"`
	ID          string       `json:"id"`
	//SourceObject *Source           `json:"-"`
	Type PaymentSourceType `json:"object"`
}

type Subscription struct {
	ID                    primitive.ObjectID             `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt             time.Time                      `json:"createdAt" bson:"createdAt"`
	DeletedAt             *time.Time                     `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt             time.Time                      `json:"updatedAt" bson:"updatedAt"`
	CreatedBy             primitive.ObjectID             `json:"createdBy" bson:"createdBy"`
	ApplicationFeePercent float64                        `json:"applicationFeePercent"`
	Billing               SubscriptionBilling            `json:"billing"`
	BillingCycleAnchor    int64                          `json:"billingCycleAnchor"`
	BillingThresholds     *SubscriptionBillingThresholds `json:"billingThresholds"`
	CancelAt              int64                          `json:"cancelAt"`
	CancelAtPeriodEnd     bool                           `json:"cancelAtPeriodEnd"`
	CanceledAt            time.Time                      `json:"canceledAt"`
	Created               int64                          `json:"created"`
	CurrentPeriodEnd      int64                          `json:"currentPeriodEnd"`
	CurrentPeriodStart    int64                          `json:"currentPeriodStart"`
	Customer              *Customer                      `json:"customer"`
	DaysUntilDue          int64                          `json:"daysUntilDue"`
	DefaultPaymentMethod  *PaymentMethod                 `json:"defaultPaymentMethod"`
	DefaultSource         *PaymentSource                 `json:"defaultSource"`
	DefaultTaxRates       []*TaxRate                     `json:"defaultTaxRates"`
	Discount              *Discount                      `json:"discount"`
	EndedAt               int64                          `json:"endedAt"`
	Items                 *SubscriptionItem              `json:"items"`
	LatestInvoice         *Invoice                       `json:"latestInvoice"`
	Livemode              bool                           `json:"livemode"`
	Metadata              map[string]string              `json:"metadata"`
	Object                string                         `json:"object"`
	//OnBehalfOf            *Account                       `json:"onBehalfOf"`
	Plan      *SubscriptionPlan  `json:"plan"`
	Quantity  int64              `json:"quantity"`
	StartDate int64              `json:"startDate"`
	Status    SubscriptionStatus `json:"status"`
	//TransferData          *SubscriptionTransferData      `json:"transferData"`
	TrialEnd   time.Time `json:"trialEnd"`
	TrialStart time.Time `json:"trialStart"`
}

// SubscriptionItemBillingThresholds is a structure representing the billing thresholds for a
// subscription item.
type SubscriptionItemBillingThresholds struct {
	UsageGTE int64 `json:"usageGte"`
}

// SubscriptionItem is the resource representing a Stripe subscription item.
// For more details see https://stripe.com/docs/api#subscription_items.
type SubscriptionItem struct {
	BillingThresholds SubscriptionItemBillingThresholds `json:"billingThresholds"`
	Created           int64                             `json:"created"`
	Deleted           bool                              `json:"deleted"`
	ID                string                            `json:"id"`
	Metadata          map[string]string                 `json:"metadata"`
	Plan              *SubscriptionPlan                 `json:"plan"`
	Quantity          int64                             `json:"quantity"`
	Subscription      string                            `json:"subscription"`
	TaxRates          []*TaxRate                        `json:"taxRates"`
}

// CreateSubscription creates new subscriptions.
func CreateSubscription(subscription Subscription) (*Subscription, error) {
	subscription.CreatedAt = time.Now()
	subscription.UpdatedAt = time.Now()
	subscription.ID = primitive.NewObjectID()
	db := database.MongoDB
	installationCollection := db.Collection(SubscriptionsCollection)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := installationCollection.InsertOne(ctx, &subscription)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("subscription.created", &subscription)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(subscription.ID.Hex(), subscription, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &subscription, nil
}

// GetSubscriptionByID gives requested subscription by id.
func GetSubscriptionByID(ID string) *Subscription {
	db := database.MongoDB
	subscription := &Subscription{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(subscription)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	id, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		log.Errorln(err)
		return subscription
	}
	filter := bson.D{{"_id", id}, {"deletedAt", bson.M{"$exists": false}}}
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	err = db.Collection(SubscriptionsCollection).FindOne(ctx, filter).Decode(&subscription)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return subscription
		}
		log.Errorln(err)
		return subscription
	}
	//set cache item
	err = cacheClient.Set(ID, subscription, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return subscription
}

// GetSubscriptions gives a list of subscriptions.
func GetSubscriptions(filter bson.D, limit int, after *string, before *string, first *int, last *int) (subscriptions []*Subscription, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(SubscriptionsCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(SubscriptionsCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		subscription := &Subscription{}
		err = cur.Decode(&subscription)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		subscriptions = append(subscriptions, subscription)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return subscriptions, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateSubscription updates the subscription.
func UpdateSubscription(p *Subscription) (*Subscription, error) {
	subscription := p
	subscription.UpdatedAt = time.Now()
	filter := bson.D{{"_id", subscription.ID}}
	db := database.MongoDB
	subscriptionsCollection := db.Collection(SubscriptionsCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := subscriptionsCollection.FindOneAndReplace(context.Background(), filter, subscription, findRepOpts).Decode(&subscription)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("subscription.updated", &subscription)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(subscription.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return subscription, nil
}

// DeleteSubscriptionByID deletes subscription by id.
func DeleteSubscriptionByID(ID string) (bool, error) {
	db := database.MongoDB
	id, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", id}}
	subscriptionsCollection := db.Collection(SubscriptionsCollection)
	res, err := subscriptionsCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("subscription.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (subscription *Subscription) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, subscription); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (subscription *Subscription) MarshalBinary() ([]byte, error) {
	return json.Marshal(subscription)
}
