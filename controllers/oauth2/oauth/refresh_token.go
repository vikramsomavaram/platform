/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package oauth

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/tribehq/platform/controllers/oauth2/util"
	"github.com/tribehq/platform/models"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

var (
	// ErrRefreshTokenNotFound ...
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	// ErrRefreshTokenExpired ...
	ErrRefreshTokenExpired = errors.New("refresh token expired")
	// ErrRequestedScopeCannotBeGreater ...
	ErrRequestedScopeCannotBeGreater = errors.New("requested scope cannot be greater")
)

// GetOrCreateRefreshToken retrieves an existing refresh token, if expired,
// the token gets deleted and new refresh token is created
func GetOrCreateRefreshToken(client *models.OAuthApplication, user *models.User, expiresIn int, scope string) (*models.RefreshToken, error) {
	// Try to fetch an existing refresh token first
	filter := bson.D{{"clientId", client.ID.Hex()}}
	if user != nil && len([]rune(user.ID.Hex())) > 0 {
		filter = append(filter, bson.E{"userId", user.ID.Hex()})
	}

	refreshToken := models.GetRefreshTokenByFilter(filter)

	// Check if the token is expired, if found
	var expired bool
	if !refreshToken.ID.IsZero() {
		expired = time.Now().UTC().After(refreshToken.ExpiresAt)
	}

	// If the refresh token has expired, delete it
	if expired {
		_, err := models.DeleteRefreshTokenByID(refreshToken.ID.Hex())
		if err != nil {
			log.Error(err)
		}
	}

	// Create a new refresh token if it expired or was not found
	if expired || refreshToken.ID.IsZero() {
		refreshToken = models.NewOAuthRefreshToken(client, user, expiresIn, scope)
		refreshToken, err := models.CreateRefreshToken(refreshToken)
		if err != nil {
			return nil, err
		}
		refreshToken.ClientID = client.ClientID
		if user != nil {
			refreshToken.UserID = user.ID.Hex()
		}
	}

	return refreshToken, nil
}

// GetValidRefreshToken returns a valid non expired refresh token
func GetValidRefreshToken(token string, client *models.OAuthApplication) (*models.RefreshToken, error) {
	// Fetch the refresh token from the database
	refreshToken := models.GetRefreshTokenByFilter(bson.D{{"clientId", client.ID}, {"token", token}})
	// Not found
	if refreshToken.ID.IsZero() {
		return nil, ErrRefreshTokenNotFound
	}
	// Check the refresh token hasn't expired
	if time.Now().UTC().After(refreshToken.ExpiresAt) {
		return nil, ErrRefreshTokenExpired
	}

	return refreshToken, nil
}

// getRefreshTokenScope returns scope for a new refresh token
func getRefreshTokenScope(refreshToken *models.RefreshToken, requestedScope string) (string, error) {
	var (
		scope = refreshToken.Scope // default to the scope originally granted by the resource owner
		err   error
	)

	// If the scope is specified in the request, get the scope string
	if requestedScope != "" {
		scope, err = GetScope(requestedScope)
		if err != nil {
			return "", err
		}
	}

	// Requested scope CANNOT include any scope not originally granted
	if !util.SpaceDelimitedStringNotGreater(scope, refreshToken.Scope) {
		return "", ErrRequestedScopeCannotBeGreater
	}

	return scope, nil
}
