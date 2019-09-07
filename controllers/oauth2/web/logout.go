package web

import (
	"github.com/labstack/echo/v4"
	"github.com/tribehq/platform/controllers/oauth2/oauth"
	"net/http"
)

func Logout(ctx echo.Context) error {
	// Get the session service from the request context
	sessionService, err := getSessionService(ctx)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	// Get the user session
	userSession, err := sessionService.GetUserSession()
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	// Delete the access and refresh tokens
	oauth.ClearUserTokens(userSession)

	// Delete the user session
	_ = sessionService.ClearUserSession()

	// Redirect back to the login page
	return redirectWithQueryString("/login", ctx.Request().URL.Query(), ctx)
}
