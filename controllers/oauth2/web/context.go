package web

import (
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/tribehq/platform/controllers/oauth2/session"
	"github.com/tribehq/platform/models"
)

type contextKey int

const (
	sessionServiceKey contextKey = 0
	clientKey         contextKey = 1
)

var (
	// ErrSessionServiceNotPresent ...
	ErrSessionServiceNotPresent = errors.New("session service not present in the request context")
	// ErrClientNotPresent ...
	ErrClientNotPresent = errors.New("client not present in the request context")
)

// Returns *session.Service from the request context
func getSessionService(ctx echo.Context) (*session.Service, error) {
	val := ctx.Get("sessionServiceKey")
	//TODO check val for nil
	//if !ok {
	//	return nil, ErrSessionServiceNotPresent
	//}

	sessionService, ok := val.(session.Service)
	if !ok {
		return nil, ErrSessionServiceNotPresent
	}

	return &sessionService, nil
}

// Returns *models.OAuthApplication from the request context
func getClient(ctx echo.Context) (*models.OAuthApplication, error) {
	val := ctx.Get("clientKey")
	//if !ok {
	//	return nil, ErrClientNotPresent
	//}

	client, ok := val.(*models.OAuthApplication)
	if !ok {
		return nil, ErrClientNotPresent
	}

	return client, nil
}
