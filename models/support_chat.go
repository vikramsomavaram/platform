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

// Chat represents a support chat message.
type Chat struct {
	ID         primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt  time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt  *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt  time.Time          `json:"updatedAt" bson:"updatedAt"`
	Name       string             `json:"name" bson:"name"`
	CreatedBy  primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	StartedBy  primitive.ObjectID `json:"startedBy" bson:"startedBy"`
	Agents     []*SupportAgent    `json:"agents" bson:"agents"`
	Messages   []*ChatMessage     `json:"messages" bson:"messages"`
	Tags       []*string          `json:"tags" bson:"tags"`
	Department *SupportDepartment `json:"department" bson:"department"`
	Rating     *string            `json:"rating" bson:"rating"`
	Comment    *string            `json:"comment" bson:"comment"`
	Notes      []*ChatNote        `json:"notes" bson:"notes"`
}

// CreateChat creates a new chat.
func CreateChat(chat Chat) (*Chat, error) {
	chat.CreatedAt = time.Now()
	chat.UpdatedAt = time.Now()
	chat.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(ChatCollection)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := collection.InsertOne(ctx, &chat)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("chat.created", &chat)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(chat.ID.Hex(), chat, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &chat, nil
}

// GetChatByID gives the requested chat by id.
func GetChatByID(ID string) (*Chat, error) {
	db := database.MongoDB
	chat := &Chat{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(chat)
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
	err = db.Collection(ChatCollection).FindOne(context.Background(), filter).Decode(&chat)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, chat, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return chat, nil
}

// GetChats gives an array of chats.
func GetChats(filter bson.D, limit int, after *string, before *string, first *int, last *int) (chats []*Chat, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(ChatCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(ChatCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		chat := &Chat{}
		err = cur.Decode(&chat)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		chats = append(chats, chat)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return chats, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateChat updates the chat.
func UpdateChat(c *Chat) (*Chat, error) {
	chat := c
	chat.UpdatedAt = time.Now()
	filter := bson.D{{"_id", chat.ID}}
	db := database.MongoDB
	chatCollection := db.Collection(ChatCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := chatCollection.FindOneAndReplace(context.Background(), filter, chat, findRepOpts).Decode(&chat)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("chat.updated", &chat)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(chat.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return chat, nil
}

// DeleteChatByID deletes the chat by id.
func DeleteChatByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	chatCollection := db.Collection(ChatCollection)
	res, err := chatCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("chat.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (chat *Chat) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, chat); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (chat *Chat) MarshalBinary() ([]byte, error) {
	return json.Marshal(chat)
}

// ChatMessage represents a chat message.
type ChatMessage struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	ChatID      string             `json:"chatId" bson:"chatId"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt   *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy   primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	Message     string             `json:"message" bson:"message"`
	AgentReadAt *time.Time         `json:"agentReadAt" bson:"agentReadAt"`
	UserReadAt  *time.Time         `json:"userReadAt" bson:"userReadAt"`
	Type        ChatMessageType    `json:"type" bson:"type"`
}

// CreateChatMessage creates new chat messages.
func CreateChatMessage(chatMessage ChatMessage) (*ChatMessage, error) {
	chatMessage.CreatedAt = time.Now()
	chatMessage.UpdatedAt = time.Now()
	chatMessage.ID = primitive.NewObjectID()
	db := database.MongoDB
	collection := db.Collection(ChatMessageCollection)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := collection.InsertOne(ctx, &chatMessage)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("chat_message.created", &chatMessage)
	cacheClient := cache.RedisClient
	//set cache item
	err = cacheClient.Set(chatMessage.ID.Hex(), chatMessage, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return &chatMessage, nil
}

// GetChatMessageByID gives the requested chat message by id.
func GetChatMessageByID(ID string) (*ChatMessage, error) {
	db := database.MongoDB
	chatMessage := &ChatMessage{}
	//try finding item in cache
	cacheClient := cache.RedisClient
	err := cacheClient.Get(ID).Scan(chatMessage)
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
	err = db.Collection(ChatMessageCollection).FindOne(context.Background(), filter).Decode(&chatMessage)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	//set cache item
	err = cacheClient.Set(ID, chatMessage, DefaultRedisCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
	return chatMessage, nil
}

// GetChatMessages gives an array of chat messages.
func GetChatMessages(filter bson.D, limit int, after *string, before *string, first *int, last *int) (chatMessages []*ChatMessage, totalCount int64, hasPrevious, hasNext bool, err error) {

	db := database.MongoDB

	tcint, filter, err := calcTotalCountWithQueryFilters(ChatMessageCollection, filter, after, before)
	pagingInfo, err := PaginationUtility(after, before, first, last, &tcint)
	if err != nil {
		return
	}
	pagingInfo.QueryOpts.SetSort(bson.M{"_id": 1})

	cur, err := db.Collection(ChatMessageCollection).Find(context.Background(), filter, &pagingInfo.QueryOpts)
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		chatMessage := &ChatMessage{}
		err = cur.Decode(&chatMessage)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Errorln(err)
		}
		chatMessages = append(chatMessages, chatMessage)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return chatMessages, totalCount, pagingInfo.HasPreviousPage, pagingInfo.HasNextPage, nil
}

// UpdateChatMessage updates the chat message.
func UpdateChatMessage(c *ChatMessage) (*ChatMessage, error) {
	chatMessage := c
	chatMessage.UpdatedAt = time.Now()
	filter := bson.D{{"_id", chatMessage.ID}}
	db := database.MongoDB
	chatMessageCollection := db.Collection(ChatMessageCollection)
	findRepOpts := &options.FindOneAndReplaceOptions{}
	findRepOpts.SetReturnDocument(options.After)
	err := chatMessageCollection.FindOneAndReplace(context.Background(), filter, chatMessage, findRepOpts).Decode(&chatMessage)
	if err != nil {
		log.Error(err)
	}
	go webhooks.NewWebhookEvent("chat_message.updated", &chatMessage)
	//Update cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(chatMessage.ID.Hex()).Err()
	if err != nil {
		log.Error(err)
	}
	return chatMessage, nil
}

// DeleteChatMessageByID deletes the chat message by id.
func DeleteChatMessageByID(ID string) (bool, error) {
	db := database.MongoDB
	oID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{{"_id", oID}}
	chatMessageCollection := db.Collection(ChatMessageCollection)
	res, err := chatMessageCollection.UpdateOne(context.Background(), filter, bson.D{{"$set", bson.D{{"deletedAt", time.Now()}}}})
	if err != nil || res.ModifiedCount < 1 {
		log.Errorln(err)
		return false, err
	}
	if res.MatchedCount < 1 {
		return false, nil
	}
	go webhooks.NewWebhookEvent("chat_message.deleted", &res)
	//Delete cache item
	cacheClient := cache.RedisClient
	err = cacheClient.Del(ID).Err()
	if err != nil {
		log.Error(err)
	}
	return true, nil
}

//UnmarshalBinary required for the redis cache to work
func (chatMessage *ChatMessage) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, chatMessage); err != nil {
		return err
	}
	return nil
}

//MarshalBinary required for the redis cache to work
func (chatMessage *ChatMessage) MarshalBinary() ([]byte, error) {
	return json.Marshal(chatMessage)
}
