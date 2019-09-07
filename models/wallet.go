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

// Wallet represents a wallet.
type Wallet struct {
	ID         primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserID     string             `json:"userId" bson:"userId"`
	CreatedAt  time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt  *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt  time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy  primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	WalletType string             `json:"walletType" bson:"walletType"` //wallet type - user wallet , driver wallet
}

// CreateWallet creates wallet.
func CreateWallet(wallet Wallet) (*Wallet, error) {
	wallet.CreatedAt = time.Now()
	wallet.UpdatedAt = time.Now()
	wallet.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(ProviderWalletTransactionsCollection)
	_, err := collection.InsertOne(context.Background(), &wallet)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("wallet.created", &wallet)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(wallet.ID.Hex(), wallet, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &wallet, nil
}

// GetWalletByID gives a wallet by id.
func GetWalletByID(ID string) (*Wallet, error) {
	db := database.MongoDB
	wallet := &Wallet{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(wallet)
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
	err = db.Collection(WalletsCollection).FindOne(context.Background(), filter).Decode(&wallet)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, wallet, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return wallet, nil
}

// GetWallets gives a list of wallets.
func GetWallets(filter bson.D, limit int, after *string, before *string, first *int, last *int) (wallets []*Wallet, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(WalletsCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(WalletsCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
ctx := context.Background()
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		wallet := &Wallet{}
		err = cur.Decode(&wallet)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		wallets = append(wallets, wallet)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return wallets, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateWallet updates wallet.
func UpdateWallet(c *Wallet) (*Wallet, error) {
	wallet := c
	wallet.UpdatedAt = time.Now()
	filter := bson.D{{"_id", wallet.ID}}
	db := database.MongoDB
	walletsCollection := db.Collection(WalletsCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := walletsCollection.FindOneAndReplace(context.Background(), filter, wallet, findRepOpts).Decode(&wallet)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("wallet.updated", &wallet)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(wallet.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return wallet, nil
}

// DeleteWalletByID deletes wallet by id.
func DeleteWalletByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	walletsCollection := db.Collection(WalletsCollection)
	res, err := walletsCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("wallet.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (wallet *Wallet) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, wallet); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (wallet *Wallet) MarshalBinary() ([]byte, error) {
	return json.Marshal(wallet)
}

// WalletTransaction represents a wallet transaction.
type WalletTransaction struct {
	ID               primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	WalletID         string             `json:"walletId" bson:"walletId"`
	CreatedAt        time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt        *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt        time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy        primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Description      string             `json:"description" bson:"description"`
	Amount           float64            `json:"amount" bson:"amount"`
	BalanceFor       BalanceFor         `json:"balanceFor" bson:"balanceFor"`
	Type             TransactionType    `json:"type" bson:"type"` //credit / debit
	RemainingBalance float64            `json:"remainingBalance" bson:"remainingBalance"`
	Metadata         interface{}        `json:"metadata" bson:"metadata"`
}

// CreateWalletTransaction creates wallet transaction.
func CreateWalletTransaction(walletTransaction WalletTransaction) (*WalletTransaction, error) {
	walletTransaction.CreatedAt = time.Now()
	walletTransaction.UpdatedAt = time.Now()
	walletTransaction.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(WalletTransactionsCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &walletTransaction)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("wallet_transaction.created", &walletTransaction)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(walletTransaction.ID.Hex(), walletTransaction, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &walletTransaction, nil
}

// GetWalletTransactionByID gives wallet transaction by id.
func GetWalletTransactionByID(ID string) (*WalletTransaction, error) {
	db := database.MongoDB
	walletTransaction := &WalletTransaction{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(walletTransaction)
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
	err = db.Collection(WalletTransactionsCollection).FindOne(context.Background(), filter).Decode(&walletTransaction)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, walletTransaction, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return walletTransaction, nil
}

// GetWalletTransactions gives a list of wallet transaction.
func GetWalletTransactions(filter bson.D, limit int, after *string, before *string, first *int, last *int) (walletTransactions []*WalletTransaction, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(WalletTransactionsCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(WalletTransactionsCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
ctx := context.Background()
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		walletTransaction := &WalletTransaction{}
		err = cur.Decode(&walletTransaction)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		walletTransactions = append(walletTransactions, walletTransaction)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return walletTransactions, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

//UnmarshalBinary required for the redis cache to work
func (walletTransaction *WalletTransaction) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, walletTransaction); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (walletTransaction *WalletTransaction) MarshalBinary() ([]byte, error) {
	return json.Marshal(walletTransaction)
}

// ProviderWalletTransaction represents a provider wallet transaction.
type ProviderWalletTransaction struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Description string             `json:"description" bson:"description"`
	Amount      float64            `json:"amount" bson:"amount"`
	BalanceFor  BalanceFor         `json:"balanceFor" bson:"balanceFor"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	CreatedBy   primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Type        TransactionType    `json:"type" bson:"type"`
	Balance     float64            `json:"balance" bson:"balance"`
}

// GetProviderWalletTransactionByID gives a provider wallet transaction by id.
func GetProviderWalletTransactionByID(ID string) (*ProviderWalletTransaction, error) {
	db := database.MongoDB
	transaction := &ProviderWalletTransaction{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(transaction)
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
	err = db.Collection(ProviderWalletTransactionsCollection).FindOne(context.Background(), filter).Decode(&transaction)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, transaction, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return transaction, nil
}

// GetProviderWalletTransactions gives a list of provider wallet transaction.
func GetProviderWalletTransactions(filter bson.D, limit int, after *string, before *string, first *int, last *int) (providerWalletTransactions []*ProviderWalletTransaction, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(ProviderWalletTransactionsCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(ProviderWalletTransactionsCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
ctx := context.Background()
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		providerWalletTransaction := &ProviderWalletTransaction{}
		err = cur.Decode(&providerWalletTransaction)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		providerWalletTransactions = append(providerWalletTransactions, providerWalletTransaction)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return providerWalletTransactions, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

//UnmarshalBinary required for the redis cache to work
func (transaction *ProviderWalletTransaction) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, transaction); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (transaction *ProviderWalletTransaction) MarshalBinary() ([]byte, error) {
	return json.Marshal(transaction)
}

// Withdrawal represents a withdrawal.
type Withdrawal struct {
	ID                     primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	BookingDate            time.Time          `json:"bookingDate" bson:"bookingDate"`
	FareAmount             float64            `json:"fareAmount" bson:"fareAmount"`
	Commission             float64            `json:"commission" bson:"commission"`
	BookingCharge          float64            `json:"bookingCharge" bson:"bookingCharge"`
	Tip                    float64            `json:"tip" bson:"tip"`
	PaymentAfterCommission float64            `json:"paymentAfterCommission" bson:"paymentAfterCommission"`
	PaymentMethod          PaymentMethod      `json:"paymentMethod" bson:"paymentMethod"`
	Invoice                Invoice            `json:"invoice" bson:"invoice"`
}

// GetWithdrawalByID gives withdrawal by id.
func GetWithdrawalByID(ID string) (*Withdrawal, error) {
	db := database.MongoDB
	withdrawal := &Withdrawal{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(withdrawal)
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
	err = db.Collection(WithdrawalsCollection).FindOne(context.Background(), filter).Decode(&withdrawal)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, withdrawal, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return withdrawal, nil
}

// GetWithdrawals gives a list of withdrawals.
func GetWithdrawals(filter bson.D, limit int, after *string, before *string, first *int, last *int) (withdrawals []*Withdrawal, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(WithdrawalsCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(WithdrawalsCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
ctx := context.Background()
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		withdrawal := &Withdrawal{}
		err = cur.Decode(&withdrawal)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		withdrawals = append(withdrawals, withdrawal)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return withdrawals, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

//UnmarshalBinary required for the redis cache to work
func (withdrawal *Withdrawal) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, withdrawal); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (withdrawal *Withdrawal) MarshalBinary() ([]byte, error) {
	return json.Marshal(withdrawal)
}
