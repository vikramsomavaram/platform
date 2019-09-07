/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package web

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/tribehq/platform/controllers/oauth2/config"
	"github.com/tribehq/platform/controllers/oauth2/oauth"
	"github.com/tribehq/platform/controllers/oauth2/session"
	"github.com/tribehq/platform/models"
	"net/http"
	"net/url"
	"strconv"
)

// ErrIncorrectResponseType a form value for response_type was not set to token or code
var ErrIncorrectResponseType = errors.New("response type not one of token or code")

func AuthorizeForm(ctx echo.Context) error {
	sessionService, client, _, responseType, _, err := authorizeCommon(ctx)
	if err != nil {
		return ctx.Render(http.StatusBadRequest, "error", map[string]interface{}{"error": err.Error()})

	}

	r := ctx.Request()
	// Render the template
	errMsg, _ := sessionService.GetFlashMessage()
	query := ctx.Request().URL.Query()
	query.Set("login_redirect_uri", r.URL.Path)
	return ctx.Render(http.StatusOK, "authorize", map[string]interface{}{
		"error":       errMsg,
		"clientID":    client.AppName,
		"scopes":      client.Scopes,
		"app_company": client.PublisherName,
		"app_url":     client.Website,
		"app_email":   client.ContactEmail,
		"queryString": getQueryString(query),
		"token":       responseType == "token",
	})
}

func Authorize(ctx echo.Context) error {
	_, client, user, responseType, redirectURI, err := authorizeCommon(ctx)
	if err != nil {
		return ctx.Render(http.StatusBadRequest, "error", map[string]interface{}{"error": err.Error()})
	}

	r := ctx.Request()

	// Get the state parameter
	state := r.Form.Get("state")

	// Has the resource owner or authorization server denied the request?
	authorized := len(r.Form.Get("allow")) > 0
	if !authorized {
		return errorRedirect(ctx, redirectURI, "access_denied", state, responseType)
	}

	// Check the requested scope
	scope, err := oauth.GetScope(r.Form.Get("scope"))
	if err != nil {
		return errorRedirect(ctx, redirectURI, "invalid_scope", state, responseType)
	}

	query := redirectURI.Query()

	// When response_type == "code", we will grant an authorization code
	if responseType == "code" {
		// Create a new authorization code
		authorizationCode, err := oauth.GrantAuthorizationCode(
			client,                  // client
			user,                    // user
			config.AuthCodeLifetime, // expires in
			redirectURI.String(),    // redirect URI
			scope,                   // scope
		)
		if err != nil {
			return errorRedirect(ctx, redirectURI, "server_error", state, responseType)

		}

		// Set query string params for the redirection URL
		query.Set("code", authorizationCode.Code)
		// Add state param if present (recommended)
		if state != "" {
			query.Set("state", state)
		}
		// And we're done here, redirect
		return redirectWithQueryString(redirectURI.String(), query, ctx)

	}

	// When response_type == "token", we will directly grant an access token
	if responseType == "token" {
		// Get access token lifetime from user input
		lifetime, err := strconv.Atoi(r.Form.Get("lifetime"))
		if err != nil {
			return errorRedirect(ctx, redirectURI, "server_error", state, responseType)
		}

		// Grant an access token
		accessToken, err := oauth.GrantAccessToken(
			client,   // client
			user,     // user
			lifetime, // expires in
			scope,    // scope
		)
		if err != nil {
			return errorRedirect(ctx, redirectURI, "server_error", state, responseType)
		}

		// Set query string params for the redirection URL
		query.Set("access_token", accessToken.AccessToken)
		query.Set("expires_in", fmt.Sprintf("%d", config.AccessTokenLifetime))
		query.Set("token_type", "Bearer")
		query.Set("scope", scope)
		// Add state param if present (recommended)
		if state != "" {
			query.Set("state", state)
		}
		// And we're done here, redirect
		return redirectWithFragment(redirectURI.String(), ctx)
	}
	return nil
}

func authorizeCommon(ctx echo.Context) (*session.Service, *models.OAuthApplication, *models.User, string, *url.URL, error) {
	// Get the session service from the request context
	sessionService, err := getSessionService(ctx)
	if err != nil {
		return nil, nil, nil, "", nil, err
	}

	// Get the client from the request context
	client, err := getClient(ctx)
	if err != nil {
		return nil, nil, nil, "", nil, err
	}

	// Get the user session
	userSession, err := sessionService.GetUserSession()
	if err != nil {
		return nil, nil, nil, "", nil, err
	}

	// Fetch the user
	user, err := oauth.FindUserByEmail(
		userSession.UserID,
	)
	if err != nil {
		return nil, nil, nil, "", nil, err
	}

	// Check the response_type is either "code" or "token"
	responseType := ctx.FormValue("response_type")
	if responseType == "" {
		ctx.QueryParam("response_type")
	}
	if responseType != "code" && responseType != "token" {
		return nil, nil, nil, "", nil, ErrIncorrectResponseType
	}

	// Fallback to the client redirect URI if not in query string
	redirectURI := ctx.FormValue("redirect_uri")

	if redirectURI == "" {
		ctx.QueryParam("redirect_uri")
	}

	if redirectURI == "" {
		redirectURI = client.RedirectURL
	}

	// // Parse the redirect URL
	parsedRedirectURI, err := url.ParseRequestURI(redirectURI)
	if err != nil {
		return nil, nil, nil, "", nil, err
	}

	return sessionService, client, user, responseType, parsedRedirectURI, nil
}
