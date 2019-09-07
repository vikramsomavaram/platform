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

// Document represents a document.
type Document struct {
	ID           primitive.ObjectID   `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt    time.Time            `json:"createdAt" bson:"createdAt"`
	DeletedAt    *time.Time           `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt    time.Time            `json:"updatedAt" bson:"updatedAt"`
	CreatedBy    primitive.ObjectID   `json:"createdBy" bson:"createdBy"`
	ExpiryDate   time.Time            `json:"expiryDate" bson:"expiryDate"`
	Name         string               `json:"name" bson:"name"`
	URL          string               `json:"url" bson:"url"`
	BelongsTo    string               `json:"belongsTo" bson:"belongsTo"`
	UploaderType DocumentUploaderType `json:"uploaderType" bson:"uploaderType"`
	IsActive     bool                 `json:"isActive" bson:"isActive"`
}

// RequiredDocument represents Manage Documents
type RequiredDocument struct {
	ID           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt    time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt    *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt    time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy    primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	DocumentFor  DocumentFor        `json:"documentFor" bson:"documentFor"`
	Country      Country            `json:"country" bson:"country"`
	ExpireOnDate bool               `json:"expireOnDate" bson:"expireOnDate"`
	DocumentName string             `json:"documentName" bson:"documentName"`
	IsActive     bool               `json:"isActive" bson:"isActive"`
}

// CreateDocument creates new documents.
func CreateDocument(document Document) (*Document, error) {
	document.CreatedAt = time.Now()
	document.UpdatedAt = time.Now()
	document.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(DocumentsCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &document)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("required_document.created", &document)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(document.ID.Hex(), document, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &document, nil
}

// GetDocumentByID gives the requested document using id.
func GetDocumentByID(ID string) (*Document, error) {
	db := database.MongoDB
	document := &Document{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(document)
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
	err = db.Collection(DocumentsCollection).FindOne(ctx, filter).Decode(&document)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, document, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return document, nil
}

// GetDocuments gives an array of documents.
func GetDocuments(filter bson.D, limit int, after *string, before *string, first *int, last *int) (documents []*Document, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(DocumentsCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(DocumentsCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
ctx := context.Background()
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		document := &Document{}
		err = cur.Decode(&document)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		documents = append(documents, document)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return documents, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateDocument updates the documents.
func UpdateDocument(c *Document) (*Document, error) {
	document := c
	document.UpdatedAt = time.Now()
	filter := bson.D{{"_id", document.ID}}
	db := database.MongoDB
	documentsCollection := db.Collection(DocumentsCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := documentsCollection.FindOneAndReplace(context.Background(), filter, document, findRepOpts).Decode(&document)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("required_document.updated", &document)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(document.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return document, nil
}

// DeleteDocumentByID deletes the document by id.
func DeleteDocumentByID(ID string) (bool, error) {
	db := database.MongoDB
	id, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", id}}
	documentsCollection := db.Collection(DocumentsCollection)
	res, err := documentsCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
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
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (document *Document) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, document); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (document *Document) MarshalBinary() ([]byte, error) {
	return json.Marshal(document)
}

// CreateRequiredDocument creates new requiredDocument.
func CreateRequiredDocument(requiredDocument RequiredDocument) (*RequiredDocument, error) {
	requiredDocument.CreatedAt = time.Now()
	requiredDocument.UpdatedAt = time.Now()
	requiredDocument.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(RequiredDocumentsCollection)
	ctx := context.Background()
	_, err := collection.InsertOne(ctx, &requiredDocument)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("required_document.created", &requiredDocument)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(requiredDocument.ID.Hex(), requiredDocument, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &requiredDocument, nil
}

// GetRequiredDocumentByID gives requiredDocument by id.
func GetRequiredDocumentByID(ID string) (*RequiredDocument, error) {
	db := database.MongoDB
	requiredDocument := &RequiredDocument{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(requiredDocument)
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
	err = db.Collection(RequiredDocumentsCollection).FindOne(ctx, filter).Decode(&requiredDocument)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, requiredDocument, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return requiredDocument, nil
}

// GetRequiredDocuments gives a list of requiredDocuments.
func GetRequiredDocuments(filter bson.D, limit int, after *string, before *string, first *int, last *int) (requiredDocuments []*RequiredDocument, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(RequiredDocumentsCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(RequiredDocumentsCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
ctx := context.Background()
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		requiredDocument := &RequiredDocument{}
		err = cur.Decode(&requiredDocument)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		requiredDocuments = append(requiredDocuments, requiredDocument)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return requiredDocuments, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateRequiredDocument updates requiredDocument.
func UpdateRequiredDocument(s *RequiredDocument) (*RequiredDocument, error) {
	requiredDocument := s
	requiredDocument.UpdatedAt = time.Now()
	filter := bson.D{{"_id", requiredDocument.ID}}
	db := database.MongoDB
	requiredDocumentsCollection := db.Collection(RequiredDocumentsCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := requiredDocumentsCollection.FindOneAndReplace(context.Background(), filter, requiredDocument, findRepOpts).Decode(&requiredDocument)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("required_document.updated", &requiredDocument)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(requiredDocument.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return requiredDocument, nil
}

// DeleteRequiredDocumentByID deletes requiredDocument by id.
func DeleteRequiredDocumentByID(ID string) (bool, error) {
	db := database.MongoDB
	id, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", id}}
	requiredDocumentsCollection := db.Collection(RequiredDocumentsCollection)
	res, err := requiredDocumentsCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil && res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("required_document.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (requiredDocument *RequiredDocument) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, requiredDocument); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (requiredDocument *RequiredDocument) MarshalBinary() ([]byte, error) {
	return json.Marshal(requiredDocument)
}
