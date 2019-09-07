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

// EnterpriseAccount represents a enterprise account.
type EnterpriseAccount struct {
	ID               primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt        time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt        *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt        time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy        primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	OrganizationName string             `json:"organizationName" bson:"organizationName"`
	OrganizationType string             `json:"organizationType" bson:"organizationType"`
	PaymentMethod    string             `json:"paymentMethod" bson:"paymentMethod"`
	Email            string             `json:"email" bson:"email"`
	Country          string             `json:"country" bson:"country"`
	State            string             `json:"state" bson:"state"`
	City             string             `json:"city" bson:"city"`
	Address          Address            `json:"address" bson:"address"`
	ZipCode          string             `json:"zipCode" bson:"zipCode"`
	Language         string             `json:"language" bson:"language"`
	PaymentBy        PaymentBy          `json:"paymentBy" bson:"paymentBy"`
	Phone            string             `json:"phone" bson:"phone"`
	IsActive         bool               `json:"isActive" bson:"isActive"`
}

// CreateEnterpriseAccount creates new enterprise accounts.
func CreateEnterpriseAccount(enterpriseAccount EnterpriseAccount) (*EnterpriseAccount, error) {
	enterpriseAccount.CreatedAt = time.Now()
	enterpriseAccount.UpdatedAt = time.Now()
	enterpriseAccount.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(EnterpriseAccountsCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &enterpriseAccount)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("enterprise_account.created", &enterpriseAccount)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(enterpriseAccount.ID.Hex(), enterpriseAccount, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &enterpriseAccount, nil
}

// GetEnterpriseAccountByID gives requested enterprise account using id.
func GetEnterpriseAccountByID(ID string) (*EnterpriseAccount, error) {
	db := database.MongoDB
	enterpriseAccount := &EnterpriseAccount{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(enterpriseAccount)
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
	err = db.Collection(EnterpriseAccountsCollection).FindOne(ctx, filter).Decode(&enterpriseAccount)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, enterpriseAccount, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return enterpriseAccount, nil
}

// GetEnterpriseAccounts gives an array of enterprise accounts.
func GetEnterpriseAccounts(filter bson.D, limit int, after *string, before *string, first *int, last *int) (enterpriseAccounts []*EnterpriseAccount, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(EnterpriseAccountsCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(EnterpriseAccountsCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
ctx := context.Background()
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		enterpriseAccount := &EnterpriseAccount{}
		err = cur.Decode(&enterpriseAccount)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		enterpriseAccounts = append(enterpriseAccounts, enterpriseAccount)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return enterpriseAccounts, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateEnterpriseAccount updates the enterprise accounts.
func UpdateEnterpriseAccount(c *EnterpriseAccount) (*EnterpriseAccount, error) {
	enterpriseAccount := c
	enterpriseAccount.UpdatedAt = time.Now()
	filter := bson.D{{"_id", enterpriseAccount.ID}}
	db := database.MongoDB
	enterpriseAccountsCollection := db.Collection(EnterpriseAccountsCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := enterpriseAccountsCollection.FindOneAndReplace(context.Background(), filter, enterpriseAccount, findRepOpts).Decode(&enterpriseAccount)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("enterprise_account.updated", &enterpriseAccount)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(enterpriseAccount.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return enterpriseAccount, nil
}

// DeleteEnterpriseAccountByID deletes the enterprise account by id.
func DeleteEnterpriseAccountByID(ID string) (bool, error) {
	db := database.MongoDB
	id, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", id}}
	enterpriseAccountsCollection := db.Collection(EnterpriseAccountsCollection)
	res, err := enterpriseAccountsCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("enterprise_account.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (enterpriseAccount *EnterpriseAccount) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, enterpriseAccount); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (enterpriseAccount *EnterpriseAccount) MarshalBinary() ([]byte, error) {
	return json.Marshal(enterpriseAccount)
}

// EnterpriseAccountPaymentReport represents enterprise account payment report.
type EnterpriseAccountPaymentReport struct {
	ID                        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt                 time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt                 *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt                 time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy                 primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	JobType                   string             `json:"jobType" bson:"jobType"`
	RideJobNo                 string             `json:"rideJobNo" bson:"rideJobNo"`
	User                      string             `json:"user" bson:"user"`
	JobDate                   string             `json:"jobDate" bson:"jobDate"`
	TotalFare                 string             `json:"totalFare" bson:"totalFare"`
	PlatformFees              string             `json:"platformFees" bson:"platformFees"`
	WalletDebit               string             `json:"walletDebit" bson:"walletDebit"`
	Tip                       string             `json:"tip" bson:"tip"`
	JobOutstandingAmount      string             `json:"jobOutstandingAmount" bson:"jobOutstandingAmount"`
	OrganizationPayAmount     string             `json:"organizationPayAmount" bson:"organizationPayAmount"`
	JobStatus                 string             `json:"jobStatus" bson:"jobStatus"`
	PaymentMethod             string             `json:"paymentMethod" bson:"paymentMethod"`
	OrganizationPaymentStatus string             `json:"organizationPaymentStatus" bson:"organizationPaymentStatus"`
	CancelledRideJobNo        string             `json:"cancelledRideJobNo" bson:"cancelledRideJobNo"`
	Organization              string             `json:"organization" bson:"organization"`
	Provider                  string             `json:"provider" bson:"provider"`
	TotalCancellationFees     string             `json:"totalCancellationFees" bson:"totalCancellationFees"`
	ProviderPaymentStatus     string             `json:"providerPaymentStatus" bson:"providerPaymentStatus"`
	IsActive                  bool               `json:"isActive" bson:"isActive"`
}

// GetEnterpriseAccountPaymentReportByID gives the request enterprise account payment report by ID
func GetEnterpriseAccountPaymentReportByID(ID string) (*EnterpriseAccountPaymentReport, error) {
	db := database.MongoDB
	report := &EnterpriseAccountPaymentReport{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(report)
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
	err = db.Collection(EnterpriseAccountPaymentReportsCollection).FindOne(ctx, filter).Decode(&report)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, report, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return report, nil
}

// GetEnterpriseAccountPaymentReports gives an array of enterprise account payment reports
func GetEnterpriseAccountPaymentReports(filter bson.D, limit int, after *string, before *string, first *int, last *int) (enterpriseAccountReports []*EnterpriseAccountPaymentReport, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(EnterpriseAccountPaymentReportsCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(EnterpriseAccountPaymentReportsCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
ctx := context.Background()
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		enterpriseAccountReport := &EnterpriseAccountPaymentReport{}
		err = cur.Decode(&enterpriseAccountReport)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		enterpriseAccountReports = append(enterpriseAccountReports, enterpriseAccountReport)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return enterpriseAccountReports, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

//UnmarshalBinary required for the redis cache to work
func (enterpriseAccountPaymentReport *EnterpriseAccountPaymentReport) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, enterpriseAccountPaymentReport); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (enterpriseAccountPaymentReport *EnterpriseAccountPaymentReport) MarshalBinary() ([]byte, error) {
	return json.Marshal(enterpriseAccountPaymentReport)
}
