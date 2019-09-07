/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package oauth

import (
	"errors"
	"github.com/tribehq/platform/models"
	"github.com/tribehq/platform/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strings"
	"time"
)

var (
	// ErrClientNotFound ...
	ErrClientNotFound = errors.New("client not found")
	// ErrInvalidClientSecret ...
	ErrInvalidClientSecret = errors.New("invalid client secret")
	// ErrClientIDTaken ...
	ErrClientIDTaken = errors.New("client ID taken")
)

// ClientExists returns true if client exists
func ClientExists(clientID string) bool {
	_, err := FindClientByClientID(clientID)
	return err == nil
}

// FindClientByClientID looks up a client by client ID
func FindClientByClientID(clientID string) (*models.OAuthApplication, error) {
	// Client IDs are case insensitive
	client := models.GetOAuthApplicationByFilter(bson.D{{"clientId", clientID}})
	if client.ID.IsZero() {
		return nil, ErrClientNotFound
	}
	return client, nil
}

// CreateClient saves a new client to database
func CreateClient(clientID, secret, redirectURI string) (*models.OAuthApplication, error) {
	return createClientCommon(clientID, secret, redirectURI)
}

// CreateClientTx saves a new client to database using injected db object
func CreateClientTx(clientID, secret, redirectURI string) (*models.OAuthApplication, error) {
	return createClientCommon(clientID, secret, redirectURI)
}

// AuthClient authenticates client
func AuthClient(clientID, secret string) (*models.OAuthApplication, error) {
	// Fetch the client
	client, err := FindClientByClientID(clientID)
	if err != nil {
		return nil, ErrClientNotFound
	}

	// Verify the secret
	if !utils.CheckPasswordHash(client.ClientSecret, secret) {
		return nil, ErrInvalidClientSecret
	}

	return client, nil
}

func createClientCommon(clientID, secret, redirectURI string) (*models.OAuthApplication, error) {
	// Check client ID
	if ClientExists(clientID) {
		return nil, ErrClientIDTaken
	}

	// Hash password
	secretHash := utils.HashPassword(secret)
	client := &models.OAuthApplication{
		ID:           primitive.NewObjectID(),
		CreatedAt:    time.Now().UTC(),
		ClientID:     strings.ToLower(clientID),
		ClientSecret: string(secretHash),
		RedirectURL:  redirectURI,
	}
	client, err := models.CreateOAuthApplication(*client)
	if err != nil {
		return nil, err
	}
	return client, nil
}
