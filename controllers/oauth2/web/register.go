/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package web

import (
	"github.com/labstack/echo/v4"
	"github.com/tribehq/platform/controllers/oauth2/oauth"
	"github.com/tribehq/platform/controllers/oauth2/oauth/roles"
	"net/http"
)

func RegisterForm(ctx echo.Context) error {
	// Get the session service from the request context
	sessionService, err := getSessionService(ctx)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	// Render the template
	errMsg, _ := sessionService.GetFlashMessage()
	return ctx.Render(http.StatusOK, "register", map[string]interface{}{
		"error":       errMsg,
		"queryString": getQueryString(ctx.Request().URL.Query()),
	})
}

func Register(ctx echo.Context) error {
	// Get the session service from the request context
	sessionService, err := getSessionService(ctx)
	if err != nil {
		return ctx.Render(http.StatusInternalServerError, "error", map[string]interface{}{"error": err.Error()})

	}

	// Check that the submitted email hasn't been registered already
	if oauth.UserExists(ctx.FormValue("email")) {
		_ = sessionService.SetFlashMessage("Email taken")
		return ctx.Redirect(http.StatusFound, ctx.Request().RequestURI)
	}

	// Create a user
	_, err = oauth.CreateUser(
		roles.User,                // role ID
		ctx.FormValue("email"),    // username
		ctx.FormValue("password"), // password
	)

	if err != nil {
		_ = sessionService.SetFlashMessage(err.Error())
		return ctx.Redirect(http.StatusFound, ctx.Request().RequestURI)
	}

	// Redirect to the login page
	return redirectWithQueryString("/login", ctx.Request().URL.Query(), ctx)
}
