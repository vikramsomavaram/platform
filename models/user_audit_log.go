/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package models

import (
	"context"
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	"github.com/tribehq/platform/lib/cache"
	"github.com/tribehq/platform/lib/database"
	"github.com/tribehq/platform/utils/webhooks"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

// UserAuditLog represents a user audit log.
type UserAuditLog struct {
	ID             primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt      time.Time          `json:"createdAt" bson:"createdAt"`
	UserID         string             `json:"userID" bson:"userID"`
	ActionType     ActionType         `json:"actionType" bson:"actionType"`
	ActionMetadata map[string]string  `json:"actionMetadata" bson:"actionMetadata"`
	ObjectID       string             `json:"objectID" bson:"objectID"`
	ObjectName     string             `json:"objectName" bson:"objectName"`
	ObjectData     interface{}        `json:"objectData" bson:"objectData"`
}

type ActionType string

const (
	Created     ActionType = "created"
	Updated     ActionType = "updated"
	Deleted     ActionType = "deleted"
	Blocked     ActionType = "blocked"
	Unblocked   ActionType = "unblocked"
	Accepted    ActionType = "accepted"
	Declined    ActionType = "declined"
	Delivered   ActionType = "delivered"
	Closed      ActionType = "closed"
	LoggedIn    ActionType = "logged in"
	LoggedOut   ActionType = "logged out"
	Activated   ActionType = "activated"
	Deactivated ActionType = "deactivated"
	Cancelled   ActionType = "cancelled"
	Requested   ActionType = "requested"
	Verified    ActionType = "verified"
	Assigned    ActionType = "assigned"
	Unassigned  ActionType = "unassigned"
	Reset       ActionType = "reset"
)

// CreateUserAuditLog creates new userAuditLog.
func CreateUserAuditLog(userAuditLog *UserAuditLog) (*UserAuditLog, error) {
	userAuditLog.CreatedAt = time.Now()
	db := database.MongoDB
	collection := db.Collection(UserAuditLogsCollection)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := collection.InsertOne(ctx, &userAuditLog)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("user_audit_log.created", &userAuditLog)
	return userAuditLog, nil
}

// GetUserAuditLogByID gives userAuditLog by id.
func GetUserAuditLogByID(ID string) (*UserAuditLog, error) {
	db := database.MongoDB
	userAuditLog := &UserAuditLog{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(userAuditLog)
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
	err = db.Collection(UserAuditLogsCollection).FindOne(ctx, filter).Decode(&userAuditLog)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, userAuditLog, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return userAuditLog, nil
}

// GetUserAuditLogs gives a list of userAuditLogs.
func GetUserAuditLogs(filter bson.D, limit int, after *string, before *string, first *int, last *int) (userAuditLogs []*UserAuditLog, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(UserAuditLogsCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(UserAuditLogsCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		userAuditLog := &UserAuditLog{}
		err = cur.Decode(&userAuditLog)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		userAuditLogs = append(userAuditLogs, userAuditLog)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return userAuditLogs, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}
