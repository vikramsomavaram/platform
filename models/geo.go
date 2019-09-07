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

// Country represents a country.
type Country struct {
	ID              primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt       time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt       *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt       time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy       primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	CountryName     string             `json:"countryName" bson:"countryName"`
	Continent       string             `json:"continent" bson:"continent"`
	Code            string             `json:"code" bson:"code"`
	PhoneCode       string             `json:"phoneCode" bson:"phoneCode"`
	DistanceUnit    string             `json:"distanceUnit" bson:"distanceUnit"`
	EmergencyNumber string             `json:"emergencyNumber" bson:"emergencyNumber"`
	Tax             map[string]string  `json:"tax" bson:"tax"`
	IsActive        bool               `json:"isActive" bson:"isActive"`
}

// State represents a state.
type State struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt   *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy   primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	CountryName string             `json:"countryName" bson:"countryName"`
	CountryCode string             `json:"countryCode" bson:"countryCode"`
	StateName   string             `json:"stateName" bson:"stateName"`
	StateCode   string             `json:"stateCode" bson:"stateCode"`
	IsActive    bool               `json:"isActive" bson:"isActive"`
}

// City represents a city.
type City struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt   *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy   primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	CountryName string             `json:"countryName" bson:"countryName"`
	CountryCode string             `json:"countryCode" bson:"countryCode"`
	StateName   string             `json:"stateName" bson:"stateName"`
	StateCode   string             `json:"stateCode" bson:"stateCode"`
	CityName    string             `json:"cityName" bson:"cityName"`
	IsActive    bool               `json:"isActive" bson:"isActive"`
}

// GetCountryByCode gives the requested country using code.
func GetCountryByCode(code string) (*Country, error) {
	db := database.MongoDB
	country := &Country{}
	filter := bson.D{{"code", code}, {"deletedAt", bson.M{"$exists": false}}}
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	err := db.Collection(CountryCollection).FindOne(ctx, filter).Decode(&country)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	return country, nil
}

// GetCountryByID gives the requested country using id.
func GetCountryByID(ID string) (*Country, error) {
	db := database.MongoDB
	country := &Country{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(country)
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
	err = db.Collection(CountryCollection).FindOne(ctx, filter).Decode(&country)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, country, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return country, nil
}

// GetCountries gives an array of countries.
func GetCountries(filter bson.D, limit int, after *string, before *string, first *int, last *int) (countries []*Country, totalCount int64, hasPrevious, hasNext bool, err error) {
	db := database.MongoDB
	tcint, filter, err := calcTotalCountWithQueryFilters(CountryCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(CountryCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		country := &Country{}
		err = cur.Decode(&country)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		countries = append(countries, country)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return countries, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// CreateCountry creates new countries.
func CreateCountry(country Country) (*Country, error) {
	country.CreatedAt = time.Now()
	country.UpdatedAt = time.Now()
	country.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(CountryCollection)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := collection.InsertOne(ctx, &country)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("country.created", &country)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(country.ID.Hex(), country, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &country, nil
}

// UpdateCountry updates the countries.
func UpdateCountry(c *Country) (*Country, error) {
	country := c
	country.UpdatedAt = time.Now()
	filter := bson.D{{"_id", country.ID}}
	db := database.MongoDB
	collection := db.Collection(CountryCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := collection.FindOneAndReplace(context.Background(), filter, country, findRepOpts).Decode(&country)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("country.updated", &country)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(country.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return country, nil
}

// DeleteCountry deletes the country.
func DeleteCountry(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	collection := db.Collection(CountryCollection)
	res, err := collection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("country.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (country *Country) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, country); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (country *Country) MarshalBinary() ([]byte, error) {
	return json.Marshal(country)
}

// GetState gives the requested state.
func GetState(countryCode string, stateCode string) (*State, error) {
	db := database.MongoDB
	state := &State{}
	filter := bson.D{{"countryCode", countryCode}, {Key: "stateCode", Value: stateCode}, {"deletedAt", bson.M{"$exists": false}}}
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	err := db.Collection(StatesCollection).FindOne(ctx, filter).Decode(&state)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	return state, nil
}

// GetStateByID gives the requested state by id.
func GetStateByID(ID string) (*State, error) {
	db := database.MongoDB
	state := &State{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(state)
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
	err = db.Collection(StatesCollection).FindOne(ctx, filter).Decode(&state)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, state, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return state, nil
}

// GetStates gives an array of states.
func GetStates(filter bson.D, limit int, after *string, before *string, first *int, last *int) (states []*State, totalCount int64, hasPrevious, hasNext bool, err error) {
	db := database.MongoDB
	tcint, filter, err := calcTotalCountWithQueryFilters(StatesCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(StatesCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		state := &State{}
		err = cur.Decode(&state)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		states = append(states, state)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return states, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// CreateState creates new states.
func CreateState(state State) (*State, error) {
	state.CreatedAt = time.Now()
	state.UpdatedAt = time.Now()
	state.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(StatesCollection)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := collection.InsertOne(ctx, &state)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("state.created", &state)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(state.ID.Hex(), state, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &state, nil
}

// UpdateState updates the states.
func UpdateState(c *State) (*State, error) {
	state := c
	state.UpdatedAt = time.Now()
	filter := bson.D{{"_id", state.ID}}
	db := database.MongoDB
	statesCollection := db.Collection(StatesCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := statesCollection.FindOneAndReplace(context.Background(), filter, state, findRepOpts).Decode(&state)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("state.updated", &state)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(state.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return state, nil
}

// DeleteState deletes the state.
func DeleteState(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	statesCollection := db.Collection(StatesCollection)
	res, err := statesCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("state.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (state *State) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, state); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (state *State) MarshalBinary() ([]byte, error) {
	return json.Marshal(state)
}

//GetCityByCode gives the requested city using code.
func GetCityByCode(code string) (*City, error) {
	db := database.MongoDB
	city := &City{}
	filter := bson.D{{"code", code}, {"deletedAt", bson.M{"$exists": false}}}
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	err := db.Collection(CitiesCollection).FindOne(ctx, filter).Decode(&city)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	return city, nil
}

// GetCityByID gives the requested city by id.
func GetCityByID(ID string) (*City, error) {
	db := database.MongoDB
	city := &City{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(city)
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
	err = db.Collection(CitiesCollection).FindOne(ctx, filter).Decode(&city)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, city, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return city, nil
}

// GetCities gives an array of cities.
func GetCities(filter bson.D, limit int, after *string, before *string, first *int, last *int) (cities []*City, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(CitiesCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(CitiesCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		city := &City{}
		err = cur.Decode(&city)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		cities = append(cities, city)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return cities, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// CreateCity creates new cities.
func CreateCity(city City) (*City, error) {
	city.CreatedAt = time.Now()
	city.UpdatedAt = time.Now()
	city.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(CitiesCollection)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := collection.InsertOne(ctx, &city)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("city.created", &city)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(city.ID.Hex(), city, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &city, nil
}

// UpdateCity updates the cities.
func UpdateCity(c *City) (*City, error) {
	city := c
	city.UpdatedAt = time.Now()
	filter := bson.D{{"_id", city.ID}}
	db := database.MongoDB
	citiesCollection := db.Collection(CitiesCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := citiesCollection.FindOneAndReplace(context.Background(), filter, city, findRepOpts).Decode(&city)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("city.updated", &city)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(city.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return city, nil
}

// DeleteCity deletes the city by id.
func DeleteCity(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	citiesCollection := db.Collection(CitiesCollection)
	res, err := citiesCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("city.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (city *City) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, city); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (city *City) MarshalBinary() ([]byte, error) {
	return json.Marshal(city)
}
