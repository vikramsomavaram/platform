/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package web

import (
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"github.com/tribehq/platform/controllers/oauth2/oauth"
	"github.com/tribehq/platform/controllers/oauth2/session"
	"net/http"
)

// guestMiddleware just initialises session
type guestMiddleware struct {
	service session.Service
}

// newGuestMiddleware creates a new guestMiddleware instance
func NewGuestMiddleware(service session.Service) *guestMiddleware {
	return &guestMiddleware{service: service}
}

// ServeHTTP as per the negroni.Handler interface
func (m *guestMiddleware) Serve() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			// Initialise the session service
			m.service.SetSessionService(ctx)
			// Attempt to start the session
			if err := m.service.StartSession(); err != nil {
				return ctx.Render(http.StatusInternalServerError, "error", map[string]interface{}{"error": err.Error()})
			}
			ctx.Set("sessionServiceKey", m.service)
			return next(ctx)
		}
	}
}

// loggedInMiddleware initialises session and makes sure the user is logged in
type loggedInMiddleware struct {
	service session.Service
}

// newLoggedInMiddleware creates a new loggedInMiddleware instance
func NewLoggedInMiddleware(service session.Service) *loggedInMiddleware {
	return &loggedInMiddleware{service: service}
}

func (m *loggedInMiddleware) Serve() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			// Initialise the session service
			m.service.SetSessionService(ctx)
			// Attempt to start the session
			if err := m.service.StartSession(); err != nil {
				return ctx.Render(http.StatusInternalServerError, "error", map[string]interface{}{"error": err.Error()})
			}

			ctx.Set("sessionServiceKey", m.service)

			// Try to get a user session
			userSession, err := m.service.GetUserSession()
			if err != nil {
				query := ctx.Request().URL.Query()
				query.Set("login_redirect_uri", ctx.Request().URL.Path)
				return redirectWithQueryString("/login", query, ctx)

			}

			// Authenticate
			if err := m.authenticate(userSession); err != nil {
				query := ctx.Request().URL.Query()
				query.Set("login_redirect_uri", ctx.Request().URL.Path)
				return redirectWithQueryString("/login", query, ctx)

			}

			// Update the user session
			err = m.service.SetUserSession(userSession)
			if err != nil {
				log.Error(err)
			}

			return next(ctx)
		}
	}
}

func (m *loggedInMiddleware) authenticate(userSession *session.UserSession) error {
	// Try to authenticate with the stored access token
	_, err := oauth.Authenticate(userSession.AccessToken)
	if err == nil {
		// Access token valid, return
		return nil
	}
	// Access token might be expired, let's try refreshing...

	// Fetch the client
	client, err := oauth.FindClientByClientID(
		userSession.ClientID, // client ID
	)
	if err != nil {
		return err
	}

	// Validate the refresh token
	theRefreshToken, err := oauth.GetValidRefreshToken(
		userSession.RefreshToken, // refresh token
		client,                   // client
	)
	if err != nil {
		return err
	}

	// Log in the user
	accessToken, refreshToken, err := oauth.Login(
		theRefreshToken.Client(),
		theRefreshToken.User(),
		theRefreshToken.Scope,
	)
	if err != nil {
		return err
	}

	userSession.AccessToken = accessToken.AccessToken
	userSession.RefreshToken = refreshToken.Token

	return nil
}

// clientMiddleware takes client_id param from the query string and
// makes a database lookup for a client with the same client ID
type clientMiddleware struct {
	service session.Service
}

// newClientMiddleware creates a new clientMiddleware instance
func NewClientMiddleware(service session.Service) *clientMiddleware {
	return &clientMiddleware{service: service}
}

func (m *clientMiddleware) Serve() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			// Fetch the client by client_id
			client, err := oauth.FindClientByClientID(ctx.QueryParam("client_id"))
			if err != nil {
				return ctx.Render(http.StatusBadRequest, "error", map[string]interface{}{"error": err.Error()})
			}
			ctx.Set("clientKey", client)
			return next(ctx)
		}
	}
}
