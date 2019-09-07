/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package models

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/tribehq/platform/lib/database"
	"github.com/tribehq/platform/utils/webhooks"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type PasswordToken struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	ExpiresAt time.Time          `json:"expiresAt" bson:"expiresAt"`
	UserID    primitive.ObjectID `json:"userID" bson:"userID"`
	Metadata  map[string]string  `json:"metadata" bson:"metadata"`
	Token     string             `json:"token" bson:"token"`
	TokenType TokenType          `json:"tokenType" bson:"tokenType"`
}

type TokenType string

const (
	ChangePassword  TokenType = "ChangePasswordToken"
	RequestPassword TokenType = "RequestPasswordToken"
)

// CreatePasswordToken creates new password token
func CreatePasswordToken(passwordToken *PasswordToken) (*PasswordToken, error) {
	passwordToken.CreatedAt = time.Now()
	passwordToken.ID = primitive.NewObjectID()
	db := database.MongoDB
	passwordTokenCollection := db.Collection(PasswordTokenCollection)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := passwordTokenCollection.InsertOne(ctx, &passwordToken)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	go webhooks.NewWebhookEvent("password_token.created", &passwordToken)
	return passwordToken, nil
}

// GetPasswordTokenByID gives password token by token
func GetPasswordToken(token string) (*PasswordToken, error) {
	db := database.MongoDB
	passwordToken := &PasswordToken{}
	filter := bson.D{{"token", token}}
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	err := db.Collection(PasswordTokenCollection).FindOne(ctx, filter).Decode(&passwordToken)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, nil
		}
		log.Errorln(err)
		return nil, err
	}
	return passwordToken, nil
}
