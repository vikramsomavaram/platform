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

// ServiceCompany represents a company.
type ServiceCompany struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Name      string             `json:"name" bson:"name"`
	Email     string             `json:"email" bson:"email"`
	Country   string             `json:"country" bson:"country"`
	State     string             `json:"state" bson:"state"`
	City      string             `json:"city" bson:"city"`
	Address   Address            `json:"address" bson:"address"`
	ZipCode   string             `json:"zipcode" bson:"zipcode"`
	Phone     string             `json:"phone" bson:"phone"`
	Language  string             `json:"language" bson:"language"`
	VATNo     string             `json:"vatNo" bson:"vatNo"`
	IsActive  bool               `json:"isActive" bson:"isActive"`
}

// CreateServiceCompany creates new company.
func CreateServiceCompany(company ServiceCompany) (*ServiceCompany, error) {
	company.CreatedAt = time.Now()
	company.UpdatedAt = time.Now()
	company.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(ServiceCompaniesCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &company)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("service_company.created", &company)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(company.ID.Hex(), company, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &company, nil
}

// GetServiceCompanyByID gives the requested company by id.
func GetServiceCompanyByID(ID string) *ServiceCompany {
	db := database.MongoDB
	company := &ServiceCompany{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(company)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}

	oID, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		log.Errorln(err)
		return company
	}
	filter := bson.D{{"_id", oID}, {"deletedAt", bson.M{"$exists": false}}}
	ctx := context.Background()
	err = db.Collection(ServiceCompaniesCollection).FindOne(ctx, filter).Decode(&company)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return company
		}
		log.Errorln(err)
		return company
	}
	//set cache item
	err = cacheClient.Set(ID, company, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return company
}

// GetServiceCompanies gives an array of companies.
func GetServiceCompanies(filter bson.D, limit int, after *string, before *string, first *int, last *int) (serviceCompanies []*ServiceCompany, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(ServiceCompaniesCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(ServiceCompaniesCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		serviceCompany := &ServiceCompany{}
		err = cur.Decode(&serviceCompany)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		serviceCompanies = append(serviceCompanies, serviceCompany)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return serviceCompanies, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateServiceCompany updates the company.
func UpdateServiceCompany(c *ServiceCompany) (*ServiceCompany, error) {
	company := c
	company.UpdatedAt = time.Now()
	filter := bson.D{{"_id", company.ID}}
	db := database.MongoDB
	companiesCollection := db.Collection(ServiceCompaniesCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := companiesCollection.FindOneAndReplace(context.Background(), filter, company, findRepOpts).Decode(&company)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("service_company.updated", &company)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(company.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return company, nil
}

// DeleteServiceCompanyByID deletes the company by id.
func DeleteServiceCompanyByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	companiesCollection := db.Collection(ServiceCompaniesCollection)
	res, err := companiesCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("service_company.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (company *ServiceCompany) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, company); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (company *ServiceCompany) MarshalBinary() ([]byte, error) {
	return json.Marshal(company)
}
