/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package oauth

import (
	"errors"
	"github.com/tribehq/platform/models"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

var (
	// ErrAuthorizationCodeNotFound ...
	ErrAuthorizationCodeNotFound = errors.New("authorization code not found")
	// ErrAuthorizationCodeExpired ...
	ErrAuthorizationCodeExpired = errors.New("authorization code expired")
)

// GrantAuthorizationCode grants a new authorization code
func GrantAuthorizationCode(client *models.OAuthApplication, user *models.User, expiresIn int, redirectURI, scope string) (*models.AuthorizationCode, error) {
	// Create a new authorization code
	authorizationCode := models.NewOAuthAuthorizationCode(client, user, expiresIn, redirectURI, scope)
	authorizationCode, err := models.CreateAuthorizationCode(authorizationCode)
	if err != nil {
		return nil, err
	}
	authorizationCode.Client = client
	authorizationCode.User = user
	return authorizationCode, nil
}

// getValidAuthorizationCode returns a valid non expired authorization code
func getValidAuthorizationCode(code, redirectURI string, client *models.OAuthApplication) (*models.AuthorizationCode, error) {
	// Fetch the auth code from the database
	authorizationCode := new(models.AuthorizationCode)
	authCode := models.GetAuthorizationCodeByFilter(bson.D{{"clientId", client.ID.Hex()}, {"code", code}})
	// Not found
	if authCode.ID.IsZero() {
		return nil, ErrAuthorizationCodeNotFound
	}

	// Redirect URI must match if it was used to obtain the authorization code
	if redirectURI != *authorizationCode.RedirectURL {
		return nil, ErrInvalidRedirectURI
	}

	// Check the authorization code hasn't expired
	if time.Now().After(authorizationCode.ExpiresAt) {
		return nil, ErrAuthorizationCodeExpired
	}

	return authorizationCode, nil
}
