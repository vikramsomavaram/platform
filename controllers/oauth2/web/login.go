/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package web

import (
	"github.com/labstack/echo/v4"
	"github.com/tribehq/platform/controllers/oauth2/oauth"
	"github.com/tribehq/platform/controllers/oauth2/session"
	"net/http"
)

func LoginForm(ctx echo.Context) error {
	// Get the session service from the request context
	sessionService, err := getSessionService(ctx)
	if err != nil {
		return ctx.Render(http.StatusInternalServerError, "error", map[string]interface{}{"error": err.Error()})
	}

	// Render the template
	errMsg, _ := sessionService.GetFlashMessage()
	return ctx.Render(http.StatusOK, "login", map[string]interface{}{
		"error":       errMsg,
		"queryString": getQueryString(ctx.Request().URL.Query()),
	})
}

func Login(ctx echo.Context) error {
	// Get the session service from the request context
	sessionService, err := getSessionService(ctx)
	if err != nil {
		return ctx.Render(http.StatusInternalServerError, "error", map[string]interface{}{"error": err.Error()})

	}

	// Get the client from the request context
	client, err := getClient(ctx)
	if err != nil {
		return ctx.Render(http.StatusBadRequest, "error", map[string]interface{}{"error": err.Error()})

	}

	// Authenticate the user
	user, err := oauth.AuthUser(
		ctx.FormValue("email"),    // username
		ctx.FormValue("password"), // password
	)
	if err != nil {
		sessionService.SetFlashMessage(err.Error())
		return ctx.Redirect(http.StatusFound, ctx.Request().RequestURI)
	}

	// Get the scope string
	scope, err := oauth.GetScope(ctx.FormValue("scope"))
	if err != nil {
		sessionService.SetFlashMessage(err.Error())
		return ctx.Redirect(http.StatusFound, ctx.Request().RequestURI)
	}

	// Log in the user
	accessToken, refreshToken, err := oauth.Login(
		client,
		user,
		scope,
	)
	if err != nil {
		sessionService.SetFlashMessage(err.Error())
		return ctx.Redirect(http.StatusFound, ctx.Request().RequestURI)
	}

	// Log in the user and store the user session in a cookie
	userSession := &session.UserSession{
		ClientID:     client.ClientID,
		UserID:       user.ID.Hex(),
		AccessToken:  accessToken.AccessToken,
		RefreshToken: refreshToken.Token,
	}

	if err := sessionService.SetUserSession(userSession); err != nil {
		sessionService.SetFlashMessage(err.Error())
		return ctx.Redirect(http.StatusFound, ctx.Request().RequestURI)
	}

	// Redirect to the authorize page by default but allow redirection to other
	// pages by specifying a path with login_redirect_uri query string param
	loginRedirectURI := ctx.Request().URL.Query().Get("login_redirect_uri")
	if loginRedirectURI == "" {
		loginRedirectURI = "https://tribe.cab/"
	}
	return redirectWithQueryString(loginRedirectURI, ctx.Request().URL.Query(), ctx)
}
