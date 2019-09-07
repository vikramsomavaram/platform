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

// BankAccountAccountHolderType is the list of allowed values for the bank account holder type.
type BankAccountAccountHolderType string

// List of values that BankAccountAccountHolderType can take.
const (
	BankAccountAccountHolderTypeCompany    BankAccountAccountHolderType = "company"
	BankAccountAccountHolderTypeIndividual BankAccountAccountHolderType = "individual"
)

// BankAccount represents a bank account.
type BankAccount struct {
	ID                    primitive.ObjectID           `json:"id,omitempty" bson:"_id,omitempty"`
	UserID                primitive.ObjectID           `json:"userID" bson:"userID"`
	CreatedAt             time.Time                    `json:"createdAt" bson:"createdAt"`
	DeletedAt             *time.Time                   `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt             time.Time                    `json:"updatedAt" bson:"updatedAt"`
	CreatedBy             primitive.ObjectID           `json:"createdBy" bson:"createdBy"`
	AccountHolderName     string                       `json:"accountHolderName" bson:"accountHolderName"`
	AccountHolderType     BankAccountAccountHolderType `json:"accountHolderType" json:"accountHolderType"`
	BankName              string                       `json:"bankName" bson:"bankName"`
	Country               string                       `json:"country" bson:"country"`
	Currency              primitive.ObjectID           `json:"currency" bson:"currency"`
	DefaultForCurrency    bool                         `json:"defaultForCurrency" bson:"defaultForCurrency"`
	Fingerprint           string                       `json:"fingerprint" bson:"fingerprint"`
	AccountNumber         string                       `json:"accountNumber" bson:"accountNumber"`
	Metadata              map[string]string            `json:"metadata" bson:"metadata"`
	RoutingNumber         string                       `json:"routingNumber" bson:"routingNumber"`
	Status                BankAccountStatus            `json:"status" bson:"status"`
	BankLocation          string                       `json:"bankLocation" bson:"bankLocation"`
	BankCountry           string                       `json:"bankCountry" bson:"bankCountry"`
	SwiftCode             string                       `json:"swiftCode" bson:"swiftCode"`
	IfscCode              string                       `json:"ifscCode" bson:"ifscCode"`
	IsVerified            bool                         `json:"isVerified" bson:"isVerified"`
	VerifiedAt            *time.Time                   `json:"verifiedAt,omitempty" bson:"verifiedAt,omitempty"`
	VerifiedBy            primitive.ObjectID           `json:"verifiedBy" bson:"verifiedBy"`
	VerificationDocuments []Document                   `json:"verificationDocuments" bson:"verificationDocuments"`
}

// CreateBankAccount creates new bank account.
func CreateBankAccount(bankAccount *BankAccount) (*BankAccount, error) {
	bankAccount.CreatedAt = time.Now()
	bankAccount.UpdatedAt = time.Now()
	bankAccount.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(BankAccountsCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &bankAccount)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(bankAccount.ID.Hex(), bankAccount, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("bank_account.created", &bankAccount)
	return bankAccount, nil
}

// GetBankAccountByID gets requested bank account by id.
func GetBankAccountByID(ID string) (*BankAccount, error) {
	db := database.MongoDB
	bankAccount := &BankAccount{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(bankAccount)
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
	err = db.Collection(BankAccountsCollection).FindOne(ctx, filter).Decode(&bankAccount)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, bankAccount, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return bankAccount, nil
}

// GetBankAccountByUserID gets requested bank account by id.
func GetBankAccountByUserID(ID string) (*BankAccount, error) {
	db := database.MongoDB
	bankAccount := &BankAccount{}
	oID, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		return nil, err
	}
	filter := bson.D{{"userID", oID}, {"deletedAt", bson.M{"$exists": false}}}
	err = db.Collection(BankAccountsCollection).FindOne(context.Background(), filter).Decode(&bankAccount)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	return bankAccount, nil
}

// GetBankAccounts gives the array of bank accounts.
func GetBankAccounts(filter bson.D, limit int, after *string, before *string, first *int, last *int) (bankAccounts []*BankAccount, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(BankAccountsCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(BankAccountsCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
ctx := context.Background()
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		bankAccount := &BankAccount{}
		err = cur.Decode(&bankAccount)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		bankAccounts = append(bankAccounts, bankAccount)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return bankAccounts, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateBankAccount updates bank account.
func UpdateBankAccount(c *BankAccount) (*BankAccount, error) {
	bankAccount := c
	bankAccount.UpdatedAt = time.Now()
	filter := bson.D{{"_id", bankAccount.ID}}
	db := database.MongoDB
	bankAccountsCollection := db.Collection(BankAccountsCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := bankAccountsCollection.FindOneAndReplace(context.Background(), filter, bankAccount, findRepOpts).Decode(&bankAccount)
	if err != nil {
		log.Error(err)
	}
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(bankAccount.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("bank_account.updated", &bankAccount)
	return bankAccount, nil
}

// DeleteBankAccountByID deletes bank account by id.
func DeleteBankAccountByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	bankAccountsCollection := db.Collection(BankAccountsCollection)
	res, err := bankAccountsCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
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
	go webhooks.NewWebhookEvent("bank_account.deleted", &res)
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (bankAccount *BankAccount) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, bankAccount); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (bankAccount *BankAccount) MarshalBinary() ([]byte, error) {
	return json.Marshal(bankAccount)
}
