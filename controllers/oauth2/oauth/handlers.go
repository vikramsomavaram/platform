/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package oauth

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/tribehq/platform/models"
	"net/http"
)

var (
	realm = "oauth2_server"

	// ErrInvalidGrantType ...
	ErrInvalidGrantType = errors.New("invalid grant type")
	// ErrInvalidClientIDOrSecret ...
	ErrInvalidClientIDOrSecret = errors.New("invalid client ID or secret")
)

// tokensHandler handles all OAuth 2.0 grant types
// (POST /v1/oauth/tokens)
func TokensHandler(ctx echo.Context) error {
	// Parse the form so r.Form becomes available
	if err := ctx.Request().ParseForm(); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Map of grant types against handler functions
	grantTypes := map[string]func(r *http.Request, client *models.OAuthApplication) (*AccessTokenResponse, error){
		"authorization_code": authorizationCodeGrant,
		"password":           passwordGrant,
		"client_credentials": clientCredentialsGrant,
		"refresh_token":      refreshTokenGrant,
	}

	// Check the grant type
	grantHandler, ok := grantTypes[ctx.Request().Form.Get("grant_type")]
	if !ok {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": ErrInvalidGrantType.Error()})
	}

	// Client auth
	client, err := basicAuthClient(ctx.Request())
	if err != nil {
		ctx.Response().Header().Set("WWW-Authenticate", fmt.Sprintf("Bearer realm=%s", realm))
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})

	}

	// Grant processing
	resp, err := grantHandler(ctx.Request(), client)
	if err != nil {
		return ctx.JSON(getErrStatusCode(err), map[string]string{"error": err.Error()})

	}

	return ctx.JSON(200, resp)

}

// introspectHandler handles OAuth 2.0 introspect request
// (POST /v1/oauth/introspect)
func IntrospectHandler(ctx echo.Context) error {
	// Client auth
	client, err := basicAuthClient(ctx.Request())
	if err != nil {
		ctx.Response().Header().Set("WWW-Authenticate", fmt.Sprintf("Bearer realm=%s", realm))
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	// Introspect the token
	resp, err := IntrospectToken(ctx.Request(), client)
	if err != nil {
		return ctx.JSON(getErrStatusCode(err), map[string]string{"error": err.Error()})
	}

	return ctx.JSON(200, resp)
}

// Get client credentials from basic auth and try to authenticate client
func basicAuthClient(r *http.Request) (*models.OAuthApplication, error) {
	// Get client credentials from basic auth
	clientID, secret, ok := r.BasicAuth()
	if !ok {
		return nil, ErrInvalidClientIDOrSecret
	}

	// Authenticate the client
	client, err := AuthClient(clientID, secret)
	if err != nil {
		// For security reasons, return a general error message
		return nil, ErrInvalidClientIDOrSecret
	}

	return client, nil
}
