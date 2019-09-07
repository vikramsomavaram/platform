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

// GeneralLabel represents a general label.
type GeneralLabel struct {
	ID                     primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt              time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt              *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt              time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy              primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Code                   string             `json:"code" bson:"code"`
	ValueInEnglishLanguage string             `json:"valueInEnglishLanguage" bson:"valueInEnglishLanguage"`
	LanguageLabel          string             `json:"languageLabel" bson:"languageLabel"`
	IsActive               bool               `json:"isActive" bson:"isActive"`
}

// CreateGeneralLabel creates new general label.
func CreateGeneralLabel(generalLabel GeneralLabel) (*GeneralLabel, error) {
	generalLabel.CreatedAt = time.Now()
	generalLabel.UpdatedAt = time.Now()
	generalLabel.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(GeneralLabelCollection)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := collection.InsertOne(ctx, &generalLabel)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("general_label.created", &generalLabel)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(generalLabel.ID.Hex(), generalLabel, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &generalLabel, nil
}

// GetGeneralLabelByID gives requested general label by id.
func GetGeneralLabelByID(ID string) (*GeneralLabel, error) {
	db := database.MongoDB
	generalLabel := &GeneralLabel{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(generalLabel)
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
	err = db.Collection(GeneralLabelCollection).FindOne(ctx, filter).Decode(&generalLabel)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, generalLabel, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return generalLabel, nil
}

// GetGeneralLabels gives a list of general labels.
func GetGeneralLabels(filter bson.D, limit int, after *string, before *string, first *int, last *int) (generalLabels []*GeneralLabel, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(GeneralLabelCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(GeneralLabelCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		generalLabel := &GeneralLabel{}
		err = cur.Decode(&generalLabel)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		generalLabels = append(generalLabels, generalLabel)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return generalLabels, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateGeneralLabel updates general label.
func UpdateGeneralLabel(c *GeneralLabel) (*GeneralLabel, error) {
	generalLabel := c
	generalLabel.UpdatedAt = time.Now()
	filter := bson.D{{"_id", generalLabel.ID}}
	db := database.MongoDB
	generalLabelCollection := db.Collection(GeneralLabelCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := generalLabelCollection.FindOneAndReplace(context.Background(), filter, generalLabel, findRepOpts).Decode(&generalLabel)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("general_label.updated", &generalLabel)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(generalLabel.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return generalLabel, nil
}

// DeleteGeneralLabelByID deletes general label by id.
func DeleteGeneralLabelByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	generalLabelCollection := db.Collection(GeneralLabelCollection)
	res, err := generalLabelCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("general_label.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (generalLabel *GeneralLabel) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, generalLabel); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (generalLabel *GeneralLabel) MarshalBinary() ([]byte, error) {
	return json.Marshal(generalLabel)
}

// FoodDeliveryLabel represents a food delivery label.
type FoodDeliveryLabel struct {
	ID                     primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt              time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt              *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt              time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy              primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Code                   string             `json:"code" bson:"code"`
	ValueInEnglishLanguage string             `json:"valueInEnglishLanguage" bson:"valueInEnglishLanguage"`
	LanguageLabel          string             `json:"languageLabel" bson:"languageLabel"`
	IsActive               bool               `json:"isActive" bson:"isActive"`
}

// CreateFoodDeliveryLabel creates new food delivery label.
func CreateFoodDeliveryLabel(foodDeliveryLabel FoodDeliveryLabel) (*FoodDeliveryLabel, error) {
	foodDeliveryLabel.CreatedAt = time.Now()
	foodDeliveryLabel.UpdatedAt = time.Now()
	foodDeliveryLabel.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(FoodDeliveryLabelCollection)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := collection.InsertOne(ctx, &foodDeliveryLabel)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("food_delivery_label.created", &foodDeliveryLabel)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(foodDeliveryLabel.ID.Hex(), foodDeliveryLabel, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &foodDeliveryLabel, nil
}

// GetFoodDeliveryLabelByID gives requested food delivery label by id.
func GetFoodDeliveryLabelByID(ID string) (*FoodDeliveryLabel, error) {
	db := database.MongoDB
	foodDeliveryLabel := &FoodDeliveryLabel{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(foodDeliveryLabel)
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
	err = db.Collection(FoodDeliveryLabelCollection).FindOne(ctx, filter).Decode(&foodDeliveryLabel)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, foodDeliveryLabel, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return foodDeliveryLabel, nil
}

// GetFoodDeliveryLabels gives a list of food delivery labels.
func GetFoodDeliveryLabels(filter bson.D, limit int, after *string, before *string, first *int, last *int) (foodDeliveryLabels []*FoodDeliveryLabel, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(FoodDeliveryLabelCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(FoodDeliveryLabelCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		foodDeliveryLabel := &FoodDeliveryLabel{}
		err = cur.Decode(&foodDeliveryLabel)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		foodDeliveryLabels = append(foodDeliveryLabels, foodDeliveryLabel)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return foodDeliveryLabels, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateFoodDeliveryLabel updates the food delivery label.
func UpdateFoodDeliveryLabel(c *FoodDeliveryLabel) (*FoodDeliveryLabel, error) {
	foodDeliveryLabel := c
	foodDeliveryLabel.UpdatedAt = time.Now()
	filter := bson.D{{"_id", foodDeliveryLabel.ID}}
	db := database.MongoDB
	foodDeliveryLabelCollection := db.Collection(FoodDeliveryLabelCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := foodDeliveryLabelCollection.FindOneAndReplace(context.Background(), filter, foodDeliveryLabel, findRepOpts).Decode(&foodDeliveryLabel)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("food_delivery_label.updated", &foodDeliveryLabel)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(foodDeliveryLabel.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return foodDeliveryLabel, nil
}

// DeleteFoodDeliveryLabelByID deletes food delivery label.
func DeleteFoodDeliveryLabelByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	foodDeliveryLabelCollection := db.Collection(FoodDeliveryLabelCollection)
	res, err := foodDeliveryLabelCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("food_delivery_label.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (foodDeliveryLabel *FoodDeliveryLabel) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, foodDeliveryLabel); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (foodDeliveryLabel *FoodDeliveryLabel) MarshalBinary() ([]byte, error) {
	return json.Marshal(foodDeliveryLabel)
}

// GroceryDeliveryLabel represents a grocery delivery label.
type GroceryDeliveryLabel struct {
	ID                     primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt              time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt              *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt              time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy              primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Code                   string             `json:"code" bson:"code"`
	ValueInEnglishLanguage string             `json:"valueInEnglishLanguage" bson:"valueInEnglishLanguage"`
	LanguageLabel          string             `json:"languageLabel" bson:"languageLabel"`
	IsActive               bool               `json:"isActive" bson:"isActive"`
}

// CreateGroceryDeliveryLabel creates new grocery delivery label.
func CreateGroceryDeliveryLabel(groceryDeliveryLabel GroceryDeliveryLabel) (*GroceryDeliveryLabel, error) {
	groceryDeliveryLabel.CreatedAt = time.Now()
	groceryDeliveryLabel.UpdatedAt = time.Now()
	groceryDeliveryLabel.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(GroceryDeliveryLabelCollection)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := collection.InsertOne(ctx, &groceryDeliveryLabel)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("grocery_delivery_label.created", &groceryDeliveryLabel)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(groceryDeliveryLabel.ID.Hex(), groceryDeliveryLabel, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &groceryDeliveryLabel, nil
}

// GetGroceryDeliveryLabelByID gives requested grocery delivery label by id.
func GetGroceryDeliveryLabelByID(ID string) (*GroceryDeliveryLabel, error) {
	db := database.MongoDB
	groceryDeliveryLabel := &GroceryDeliveryLabel{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(groceryDeliveryLabel)
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
	err = db.Collection(GroceryDeliveryLabelCollection).FindOne(ctx, filter).Decode(&groceryDeliveryLabel)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, groceryDeliveryLabel, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return groceryDeliveryLabel, nil
}

// GetGroceryDeliveryLabels gives a list of grocery delivery label.
func GetGroceryDeliveryLabels(filter bson.D, limit int, after *string, before *string, first *int, last *int) (groceryDeliveryLabels []*GroceryDeliveryLabel, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(GroceryDeliveryLabelCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(GroceryDeliveryLabelCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		groceryDeliveryLabel := &GroceryDeliveryLabel{}
		err = cur.Decode(&groceryDeliveryLabel)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		groceryDeliveryLabels = append(groceryDeliveryLabels, groceryDeliveryLabel)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return groceryDeliveryLabels, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateGroceryDeliveryLabel updates grocery delivery label.
func UpdateGroceryDeliveryLabel(c *GroceryDeliveryLabel) (*GroceryDeliveryLabel, error) {
	groceryDeliveryLabel := c
	groceryDeliveryLabel.UpdatedAt = time.Now()
	filter := bson.D{{"_id", groceryDeliveryLabel.ID}}
	db := database.MongoDB
	groceryDeliveryLabelCollection := db.Collection(GroceryDeliveryLabelCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := groceryDeliveryLabelCollection.FindOneAndReplace(context.Background(), filter, groceryDeliveryLabel, findRepOpts).Decode(&groceryDeliveryLabel)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("grocery_delivery_label.updated", &groceryDeliveryLabel)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(groceryDeliveryLabel.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return groceryDeliveryLabel, nil
}

// DeleteGroceryDeliveryLabelByID deletes grocery delivery label by id.
func DeleteGroceryDeliveryLabelByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	groceryDeliveryLabelCollection := db.Collection(GroceryDeliveryLabelCollection)
	res, err := groceryDeliveryLabelCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("grocery_delivery_label.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (groceryDeliveryLabel *GroceryDeliveryLabel) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, groceryDeliveryLabel); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (groceryDeliveryLabel *GroceryDeliveryLabel) MarshalBinary() ([]byte, error) {
	return json.Marshal(groceryDeliveryLabel)
}

// WineDeliveryLabel represents a wine delivery label.
type WineDeliveryLabel struct {
	ID                     primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt              time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt              *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt              time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy              primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Code                   string             `json:"code" bson:"code"`
	ValueInEnglishLanguage string             `json:"valueInEnglishLanguage" bson:"valueInEnglishLanguage"`
	LanguageLabel          string             `json:"languageLabel" bson:"languageLabel"`
	IsActive               bool               `json:"isActive" bson:"isActive"`
}

// CreateWineDeliveryLabel creates new wine delivery label.
func CreateWineDeliveryLabel(wineDeliveryLabel WineDeliveryLabel) (*WineDeliveryLabel, error) {
	wineDeliveryLabel.CreatedAt = time.Now()
	wineDeliveryLabel.UpdatedAt = time.Now()
	wineDeliveryLabel.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(WineDeliveryLabelCollection)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := collection.InsertOne(ctx, &wineDeliveryLabel)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("wine_delivery_label.created", &wineDeliveryLabel)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(wineDeliveryLabel.ID.Hex(), wineDeliveryLabel, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &wineDeliveryLabel, nil
}

// GetWineDeliveryLabelByID gives requested wine delivery label by id.
func GetWineDeliveryLabelByID(ID string) (*WineDeliveryLabel, error) {
	db := database.MongoDB
	wineDeliveryLabel := &WineDeliveryLabel{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(wineDeliveryLabel)
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
	err = db.Collection(WineDeliveryLabelCollection).FindOne(context.Background(), filter).Decode(&wineDeliveryLabel)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, wineDeliveryLabel, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return wineDeliveryLabel, nil
}

// GetWineDeliveryLabels gives a list of wine delivery labels.
func GetWineDeliveryLabels(filter bson.D, limit int, after *string, before *string, first *int, last *int) (wineDeliveryLabels []*WineDeliveryLabel, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(WineDeliveryLabelCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(WineDeliveryLabelCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		wineDeliveryLabel := &WineDeliveryLabel{}
		err = cur.Decode(&wineDeliveryLabel)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		wineDeliveryLabels = append(wineDeliveryLabels, wineDeliveryLabel)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return wineDeliveryLabels, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateWineDeliveryLabel updates wine delivery label.
func UpdateWineDeliveryLabel(c *WineDeliveryLabel) (*WineDeliveryLabel, error) {
	wineDeliveryLabel := c
	wineDeliveryLabel.UpdatedAt = time.Now()
	filter := bson.D{{"_id", wineDeliveryLabel.ID}}
	db := database.MongoDB
	wineDeliveryLabelCollection := db.Collection(WineDeliveryLabelCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := wineDeliveryLabelCollection.FindOneAndReplace(context.Background(), filter, wineDeliveryLabel, findRepOpts).Decode(&wineDeliveryLabel)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("wine_delivery_label.updated", &wineDeliveryLabel)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(wineDeliveryLabel.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return wineDeliveryLabel, nil
}

// DeleteWineDeliveryLabelByID deletes wine delivery label by id.
func DeleteWineDeliveryLabelByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	wineDeliveryLabelCollection := db.Collection(WineDeliveryLabelCollection)
	res, err := wineDeliveryLabelCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("wine_delivery_label.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (wineDeliveryLabel *WineDeliveryLabel) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, wineDeliveryLabel); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (wineDeliveryLabel *WineDeliveryLabel) MarshalBinary() ([]byte, error) {
	return json.Marshal(wineDeliveryLabel)
}
