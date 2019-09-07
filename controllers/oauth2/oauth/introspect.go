/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package oauth

import (
	"errors"
	"github.com/tribehq/platform/models"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
)

const (
	// AccessTokenHint ...
	AccessTokenHint = "access_token"
	// RefreshTokenHint ...
	RefreshTokenHint = "refresh_token"
)

var (
	// ErrTokenMissing ...
	ErrTokenMissing = errors.New("token missing")
	// ErrTokenHintInvalid ...
	ErrTokenHintInvalid = errors.New("invalid token hint")
)

func IntrospectToken(r *http.Request, client *models.OAuthApplication) (*IntrospectResponse, error) {
	// Parse the form so r.Form becomes available
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	// Get token from the query
	token := r.Form.Get("token")
	if token == "" {
		return nil, ErrTokenMissing
	}

	// Get token type hint from the query
	tokenTypeHint := r.Form.Get("token_type_hint")

	// Default to access token hint
	if tokenTypeHint == "" {
		tokenTypeHint = AccessTokenHint
	}

	switch tokenTypeHint {
	case AccessTokenHint:
		accessToken, err := Authenticate(token)
		if err != nil {
			return nil, err
		}
		return NewIntrospectResponseFromAccessToken(accessToken)
	case RefreshTokenHint:
		refreshToken, err := GetValidRefreshToken(token, client)
		if err != nil {
			return nil, err
		}
		return NewIntrospectResponseFromRefreshToken(refreshToken)
	default:
		return nil, ErrTokenHintInvalid
	}
}

// NewIntrospectResponseFromAccessToken ...
func NewIntrospectResponseFromAccessToken(accessToken *models.AccessToken) (*IntrospectResponse, error) {
	var introspectResponse = &IntrospectResponse{
		Active:    true,
		Scope:     accessToken.Scope,
		TokenType: Bearer,
		ExpiresAt: int(accessToken.ExpiresAt.Unix()),
	}

	if accessToken.ClientID != "" {
		client := models.GetOAuthApplicationByFilter(bson.D{{"clientId", accessToken.ClientID}})
		if client.ID.IsZero() {
			return nil, ErrClientNotFound
		}
	}

	if accessToken.UserID != "" {
		user := models.GetUserByID(accessToken.UserID)
		if user.ID.IsZero() {
			return nil, ErrUserNotFound
		}
	}

	return introspectResponse, nil
}

// NewIntrospectResponseFromRefreshToken ...
func NewIntrospectResponseFromRefreshToken(refreshToken *models.RefreshToken) (*IntrospectResponse, error) {
	var introspectResponse = &IntrospectResponse{
		Active:    true,
		Scope:     refreshToken.Scope,
		TokenType: Bearer,
		ExpiresAt: int(refreshToken.ExpiresAt.Unix()),
	}

	if refreshToken.ClientID != "" {
		client := models.GetOAuthApplicationByFilter(bson.D{{"clientId", refreshToken.ClientID}})
		if client.ID.IsZero() {
			return nil, ErrClientNotFound
		}
	}

	if refreshToken.UserID != "" {
		user := models.GetUserByID(refreshToken.UserID)
		if user.ID.IsZero() {
			return nil, ErrUserNotFound
		}
	}

	return introspectResponse, nil
}
