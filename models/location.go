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

//GeoFenceLocation represents a geo fenced location.
type GeoFenceLocation struct {
	ID           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt    time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt    *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt    time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy    primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	LocationType string             `json:"locationType" bson:"locationType"`
	IsActive     bool               `json:"isActive" bson:"isActive"`
	Name         string             `json:"name" bson:"name"`
	Country      string             `json:"country" bson:"country"`
	GeoJSON      string             `json:"geoJson" bson:"geoJson"`
	LocationFor  string             `json:"locationFor" bson:"locationFor"`
}

// CreateGeoFenceLocation creates new geo fenced location.
func CreateGeoFenceLocation(location GeoFenceLocation) (*GeoFenceLocation, error) {
	location.CreatedAt = time.Now()
	location.UpdatedAt = time.Now()
	location.ID = primitive.NewObjectID()
	db := database.MongoDB
	installationCollection := db.Collection(GeoFenceLocationCollection)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := installationCollection.InsertOne(ctx, &location)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("geo_fence_location.created", &location)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(location.ID.Hex(), location, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &location, nil
}

// GetGeoFenceLocationByID gives requested geo fenced location by id.
func GetGeoFenceLocationByID(ID string) (*GeoFenceLocation, error) {
	db := database.MongoDB
	location := &GeoFenceLocation{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(location)
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
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	err = db.Collection(GeoFenceLocationCollection).FindOne(ctx, filter).Decode(&location)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, location, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return location, nil
}

// GetGeoFenceLocations gives a list of geo fence location.
func GetGeoFenceLocations(filter bson.D, limit int, after *string, before *string, first *int, last *int) (geoFenceLocations []*GeoFenceLocation, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(GeoFenceLocationCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(GeoFenceLocationCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		geoFenceLocation := &GeoFenceLocation{}
		err = cur.Decode(&geoFenceLocation)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		geoFenceLocations = append(geoFenceLocations, geoFenceLocation)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return geoFenceLocations, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateGeoFenceLocation updates geo fence location.
func UpdateGeoFenceLocation(location *GeoFenceLocation) (*GeoFenceLocation, error) {
	location.UpdatedAt = time.Now()
	filter := bson.D{{"_id", location.ID}}
	db := database.MongoDB
	geoFenceRestrictedAreaCollection := db.Collection(GeoFenceLocationCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := geoFenceRestrictedAreaCollection.FindOneAndReplace(context.Background(), filter, location, findRepOpts).Decode(&location)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("geo_fence_location.updated", &location)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(location.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return location, nil
}

// DeleteGeoFenceLocationByID deletes geo fence location by id.
func DeleteGeoFenceLocationByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	geoFenceRestrictedAreaCollection := db.Collection(GeoFenceLocationCollection)
	res, err := geoFenceRestrictedAreaCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	log.Warningln(res)
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("geo_fence_location.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (location *GeoFenceLocation) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, location); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (location *GeoFenceLocation) MarshalBinary() ([]byte, error) {
	return json.Marshal(location)
}

//UserLocation represents a users current location.
type UserLocation struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`
	UserID    string             `json:"userID" bson:"userID"`
	Location  Location           `json:"location" bson:"location"`
}

// Location represents a location.
type Location struct {
	Type        string    `json:"type" bson:"type"`
	Coordinates []float64 `json:"coordinates" bson:"coordinates"`
}

// CreateUserLocation creates new user location.
func CreateUserLocation(userLocation *UserLocation) (*UserLocation, error) {
	userLocation.CreatedAt = time.Now()
	userLocation.UpdatedAt = time.Now()
	userLocation.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(UserLocationLogCollection)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := collection.InsertOne(ctx, &userLocation)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("user_location.created", &userLocation)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(userLocation.ID.Hex(), userLocation, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return userLocation, nil
}

// GetUserLocationByID gives a user location by id.
func GetUserLocationByID(ID string) (*UserLocation, error) {
	db := database.MongoDB
	userLocation := &UserLocation{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(userLocation)
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
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	err = db.Collection(UserLocationLogCollection).FindOne(ctx, filter).Decode(&userLocation)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, userLocation, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return userLocation, nil
}

// GetUserLocations gives a list of user locations.
func GetUserLocations(filter bson.D, limit int, after *string, before *string, first *int, last *int) (userLocations []*UserLocation, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(UserLocationLogCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(UserLocationLogCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		userLocation := &UserLocation{}
		err = cur.Decode(&userLocation)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		userLocations = append(userLocations, userLocation)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return userLocations, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateUserLocation updates user location.
func UpdateUserLocation(c *UserLocation) (*UserLocation, error) {
	userLocation := c
	userLocation.UpdatedAt = time.Now()
	filter := bson.D{{"_id", userLocation.ID}}
	db := database.MongoDB
	userLocationLogCollection := db.Collection(UserLocationLogCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := userLocationLogCollection.FindOneAndReplace(context.Background(), filter, userLocation, findRepOpts).Decode(&userLocation)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("user_location.updated", &userLocation)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(userLocation.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return userLocation, nil
}

// DeleteUserLocationByID deletes user location by id.
func DeleteUserLocationByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	userLocationLogCollection := db.Collection(UserLocationLogCollection)
	res, err := userLocationLogCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("user_location.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (userLocation *UserLocation) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, userLocation); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (userLocation *UserLocation) MarshalBinary() ([]byte, error) {
	return json.Marshal(userLocation)
}

//ServiceProviderLocation Driver's current location
type ServiceProviderLocation struct {
	ID                primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	ServiceProviderID string             `json:"serviceProviderID" bson:"serviceProviderID"`
	CreatedAt         time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt         *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt         time.Time          `json:"updatedAt" bson:"updatedAt"`
	Location          Location           `json:"location" bson:"location"`
}

// CreateServiceProviderLocation creates new service provider location.
func CreateServiceProviderLocation(serviceProviderLocation *ServiceProviderLocation) (*ServiceProviderLocation, error) {
	serviceProviderLocation.CreatedAt = time.Now()
	serviceProviderLocation.UpdatedAt = time.Now()
	db := database.MongoDB
	collection := db.Collection(ServiceProviderLocationCollection)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := collection.InsertOne(ctx, &serviceProviderLocation)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("service_provider_location.created", &serviceProviderLocation)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(serviceProviderLocation.ID.Hex(), serviceProviderLocation, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return serviceProviderLocation, nil
}

// GetServiceProviderLocationByID gives service provider by id.
func GetServiceProviderLocationByID(ID string) (*ServiceProviderLocation, error) {
	db := database.MongoDB
	serviceProviderLocation := &ServiceProviderLocation{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(serviceProviderLocation)
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
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	err = db.Collection(ServiceProviderLocationCollection).FindOne(ctx, filter).Decode(&serviceProviderLocation)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, serviceProviderLocation, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return serviceProviderLocation, nil
}

// GetServiceProviderLocations gives a list of provider locations.
func GetServiceProviderLocations(filter bson.D, limit int, after *string, before *string, first *int, last *int) (serviceProviderLocations []*ServiceProviderLocation, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(ServiceProviderLocationCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(ServiceProviderLocationCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		serviceProviderLocation := &ServiceProviderLocation{}
		err = cur.Decode(&serviceProviderLocation)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		serviceProviderLocations = append(serviceProviderLocations, serviceProviderLocation)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return serviceProviderLocations, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateServiceProviderLocation updates provider location.
func UpdateServiceProviderLocation(c *ServiceProviderLocation) (*ServiceProviderLocation, error) {
	serviceProviderLocation := c
	serviceProviderLocation.UpdatedAt = time.Now()
	filter := bson.D{{"_id", serviceProviderLocation.ID}}
	db := database.MongoDB
	serviceProviderLocationsCollection := db.Collection(ServiceProviderLocationCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := serviceProviderLocationsCollection.FindOneAndReplace(context.Background(), filter, serviceProviderLocation, findRepOpts).Decode(&serviceProviderLocation)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("service_provider_location.updated", &serviceProviderLocation)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(serviceProviderLocation.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return serviceProviderLocation, nil
}

// DeleteServiceProviderLocationByID deletes provider location.
func DeleteServiceProviderLocationByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	serviceProviderLocationsCollection := db.Collection(ServiceProviderLocationCollection)
	res, err := serviceProviderLocationsCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("service_provider_location.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (serviceProviderLocation *ServiceProviderLocation) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, serviceProviderLocation); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (serviceProviderLocation *ServiceProviderLocation) MarshalBinary() ([]byte, error) {
	return json.Marshal(serviceProviderLocation)
}

// GeoFenceRestrictedArea represents a geo fence restriction area.
type GeoFenceRestrictedArea struct {
	ID              primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt       time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt       *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt       time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy       primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Address         string             `json:"address" bson:"address"`
	Area            string             `json:"area" bson:"area"`
	RestrictType    RestrictType       `json:"restrictType" bson:"restrictType"`
	RestrictArea    RestrictArea       `json:"restrictArea" bson:"restrictArea"`
	GeoLocationArea string             `json:"geoLocationArea" bson:"geoLocationArea"`
	IsActive        bool               `json:"isActive" bson:"isActive"`
}

// CreateGeoFenceRestrictedArea creates geo fence restricted area.
func CreateGeoFenceRestrictedArea(geoFenceRestrictedArea GeoFenceRestrictedArea) (*GeoFenceRestrictedArea, error) {
	geoFenceRestrictedArea.CreatedAt = time.Now()
	geoFenceRestrictedArea.UpdatedAt = time.Now()
	geoFenceRestrictedArea.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(GeoFenceRestrictedAreaCollection)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := collection.InsertOne(ctx, &geoFenceRestrictedArea)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("geo_fence_restricted_area.created", &geoFenceRestrictedArea)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(geoFenceRestrictedArea.ID.Hex(), geoFenceRestrictedArea, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &geoFenceRestrictedArea, nil
}

// GetGeoFenceRestrictedAreaByID gives geo fence restricted area by id.
func GetGeoFenceRestrictedAreaByID(ID string) (*GeoFenceRestrictedArea, error) {
	db := database.MongoDB
	geoFenceRestrictedArea := &GeoFenceRestrictedArea{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(geoFenceRestrictedArea)
	if err != nil && err != redis.Nil {
		log.Error(err)
	} else if err == redis.Nil {
		//key is empty or not set
	}
	oID, _ := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}, {"deletedAt", bson.M{"$exists": false}}}
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	err = db.Collection(GeoFenceRestrictedAreaCollection).FindOne(ctx, filter).Decode(&geoFenceRestrictedArea)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, geoFenceRestrictedArea, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return geoFenceRestrictedArea, nil
}

// GetGeoFenceRestrictedAreas gives a list of geo fence restricted areas.
func GetGeoFenceRestrictedAreas(filter bson.D, limit int, after *string, before *string, first *int, last *int) (geoFenceRestrictedAreas []*GeoFenceRestrictedArea, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(GeoFenceRestrictedAreaCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(GeoFenceRestrictedAreaCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		geoFenceRestrictedArea := &GeoFenceRestrictedArea{}
		err = cur.Decode(&geoFenceRestrictedArea)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		geoFenceRestrictedAreas = append(geoFenceRestrictedAreas, geoFenceRestrictedArea)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return geoFenceRestrictedAreas, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateGeoFenceRestrictedArea updates geo fence restricted area.
func UpdateGeoFenceRestrictedArea(c *GeoFenceRestrictedArea) (*GeoFenceRestrictedArea, error) {
	geoFenceRestrictedArea := c
	geoFenceRestrictedArea.UpdatedAt = time.Now()
	filter := bson.D{{"_id", geoFenceRestrictedArea.ID}}
	db := database.MongoDB
	geoFenceRestrictedAreaCollection := db.Collection(GeoFenceRestrictedAreaCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := geoFenceRestrictedAreaCollection.FindOneAndReplace(context.Background(), filter, geoFenceRestrictedArea, findRepOpts).Decode(&geoFenceRestrictedArea)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("geo_fence_restricted_area.updated", &geoFenceRestrictedArea)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(geoFenceRestrictedArea.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return geoFenceRestrictedArea, nil
}

// DeleteGeoFenceRestrictedAreaByID deletes geo fence restricted area.
func DeleteGeoFenceRestrictedAreaByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	geoFenceRestrictedAreaCollection := db.Collection(GeoFenceRestrictedAreaCollection)
	res, err := geoFenceRestrictedAreaCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("geo_fence_restricted_area.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (geoFenceRestrictedArea *GeoFenceRestrictedArea) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, geoFenceRestrictedArea); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (geoFenceRestrictedArea *GeoFenceRestrictedArea) MarshalBinary() ([]byte, error) {
	return json.Marshal(geoFenceRestrictedArea)
}
