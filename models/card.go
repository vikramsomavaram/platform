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

// Card represents a card.
type Card struct {
	ID             primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt      time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt      *time.Time         `json:"-,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt      time.Time          `json:"updatedAt" bson:"updatedAt"`
	AddressCity    string             `json:"addressCity" bson:"addressCity"`
	AddressCountry string             `json:"addressCountry" bson:"addressCountry"`
	AddressLine1   string             `json:"addressLine1" bson:"addressLine1"`
	//AddressLine1Check  CardVerification `json:"address_line1_check"`
	AddressLine2 string `json:"addressLine2" bson:"addressLine2"`
	AddressState string `json:"addressState" bson:"addressState"`
	AddressZip   string `json:"addressZip" bson:"addressZip"`
	//AddressZipCheck    CardVerification `json:"address_zip_check"`
	//Brand              CardBrand        `json:"brand"`
	//CVCCheck           CardVerification `json:"cvc_check"`
	Country            string   `json:"country" bson:"country"`
	Currency           Currency `json:"currency" bson:"currency"`
	User               *User    `json:"user" bson:"user"`
	DefaultForCurrency bool     `json:"defaultForCurrency" bson:"defaultForCurrency"`
	Deleted            bool     `json:"deleted" bson:"deleted"`

	// Description is a succinct summary of the card's information.
	//
	// Please note that this field is for internal use only and is not returned
	// as part of standard API requests.
	Description string `json:"description"`

	DynamicLast4 string `json:"dynamicLast4"`
	ExpMonth     uint8  `json:"expMonth"`
	ExpYear      uint16 `json:"expYear"`
	Fingerprint  string `json:"fingerprint"`
	//Funding      CardFunding `json:"funding"`

	// IIN is the card's "Issuer Identification Number".
	//
	// Please note that this field is for internal use only and is not returned
	// as part of standard API requests.
	IIN string `json:"iin"`

	// Issuer is a bank or financial institution that provides the card.
	//
	// Please note that this field is for internal use only and is not returned
	// as part of standard API requests.
	Issuer string `json:"issuer"`

	Last4    string `json:"last4"`
	Metadata string `json:"metadata"`
	Name     string `json:"name"`
	//Recipient          *Recipient             `json:"recipient"`
	//ThreeDSecure       *ThreeDSecure          `json:"three_d_secure"`
	//TokenizationMethod CardTokenizationMethod `json:"tokenization_method"`
}

// CreateCard creates new cards.
func CreateCard(card Card) (*Card, error) {
	card.CreatedAt = time.Now()
	card.UpdatedAt = time.Now()
	card.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(CardsCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &card)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("card.created", &card)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(card.ID.Hex(), card, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &card, nil
}

// GetCardByID gives the requested card by id.
func GetCardByID(ID string) (*Card, error) {
	db := database.MongoDB
	card := &Card{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(card)
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
	err = db.Collection(CardsCollection).FindOne(ctx, filter).Decode(&card)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, card, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return card, nil
}

// GetCards gives an array of cards.
func GetCards(filter bson.D, limit int, after *string, before *string, first *int, last *int) (cards []*Card, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(CardsCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(CardsCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		card := &Card{}
		err = cur.Decode(&card)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		cards = append(cards, card)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return cards, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateCard updates the card.
func UpdateCard(c *Card) (*Card, error) {
	card := c
	card.UpdatedAt = time.Now()
	filter := bson.D{{"_id", card.ID}}
	db := database.MongoDB
	cardsCollection := db.Collection(CardsCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := cardsCollection.FindOneAndReplace(context.Background(), filter, card, findRepOpts).Decode(&card)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("card.updated", &card)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(card.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return card, nil
}

// DeleteCardByID deletes the card by id.
func DeleteCardByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	cardsCollection := db.Collection(CardsCollection)
	res, err := cardsCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("card.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (card *Card) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, card); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (card *Card) MarshalBinary() ([]byte, error) {
	return json.Marshal(card)
}
