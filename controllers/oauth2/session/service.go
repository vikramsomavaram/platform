/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package session

import (
	"encoding/gob"
	"errors"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/tribehq/platform/controllers/oauth2/config"
)

// Service wraps session functionality
type Service struct {
	sessionStore   sessions.Store
	sessionOptions *sessions.Options
	session        *sessions.Session
	ctx            echo.Context
}

// UserSession has user data stored in a session after logging in
type UserSession struct {
	ClientID     string
	UserID       string
	AccessToken  string
	RefreshToken string
}

var (
	// StorageSessionName ...
	StorageSessionName = "oauth2_server_session"
	// UserSessionKey ...
	UserSessionKey = "oauth2_server_user"
	// ErrSessonNotStarted ...
	ErrSessonNotStarted = errors.New("session not started")
)

func init() {
	// Register a new datatype for storage in sessions
	gob.Register(new(UserSession))
}

// NewService returns a new Service instance
func NewService(sessionStore sessions.Store) *Service {
	return &Service{
		// Session cookie storage
		sessionStore: sessionStore,
		// Session options
		sessionOptions: &sessions.Options{
			Path:     config.SessionCookiePath,
			MaxAge:   config.SessionCookieMaxAge,
			HttpOnly: config.SessionCookieHTTPOnly,
		},
	}
}

// SetSessionService sets the request and responseWriter on the session service
func (s *Service) SetSessionService(ctx echo.Context) {
	s.ctx = ctx
}

// SetSessionSave saves to the session
func (s *Service) SetSessionSave(ctx echo.Context) error {
	return s.session.Save(ctx.Request(), ctx.Response())
}

// StartSession starts a new session. This method must be called before other
// public methods of this struct as it sets the internal session object
func (s *Service) StartSession() error {
	sess, err := session.Get(StorageSessionName, s.ctx)
	if err != nil {
		return err
	}
	s.session = sess
	return nil
}

// GetUserSession returns the user session
func (s *Service) GetUserSession() (*UserSession, error) {
	// Make sure StartSession has been called
	if s.session == nil {
		return nil, ErrSessonNotStarted
	}

	// Retrieve our user session struct and type-assert it
	userSession, ok := s.session.Values[UserSessionKey].(*UserSession)
	if !ok {
		return nil, errors.New("user session type assertion error")
	}

	return userSession, nil
}

// SetUserSession saves the user session
func (s *Service) SetUserSession(userSession *UserSession) error {
	// Make sure StartSession has been called
	if s.session == nil {
		return ErrSessonNotStarted
	}

	// Set a new user session
	s.session.Values[UserSessionKey] = userSession
	return s.session.Save(s.ctx.Request(), s.ctx.Response())
}

// ClearUserSession deletes the user session
func (s *Service) ClearUserSession() error {
	// Make sure StartSession has been called
	if s.session == nil {
		return ErrSessonNotStarted
	}

	// Delete the user session
	delete(s.session.Values, UserSessionKey)
	return s.session.Save(s.ctx.Request(), s.ctx.Response())
}

// SetFlashMessage sets a flash message,
// useful for displaying an error after 302 redirection
func (s *Service) SetFlashMessage(msg string) error {
	// Make sure StartSession has been called
	if s.session == nil {
		return ErrSessonNotStarted
	}

	// Add the flash message
	s.session.AddFlash(msg)
	return s.session.Save(s.ctx.Request(), s.ctx.Response())
}

// GetFlashMessage returns the first flash message
func (s *Service) GetFlashMessage() (interface{}, error) {
	// Make sure StartSession has been called
	if s.session == nil {
		return nil, ErrSessonNotStarted
	}

	// Get the last flash message from the stack
	if flashes := s.session.Flashes(); len(flashes) > 0 {
		// We need to save the session, otherwise the flash message won't be removed
		_ = s.session.Save(s.ctx.Request(), s.ctx.Response())
		return flashes[0], nil
	}

	// No flash messages in the stack
	return nil, nil
}
