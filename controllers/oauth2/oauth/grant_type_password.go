package oauth

import (
	"errors"
	"github.com/tribehq/platform/controllers/oauth2/config"
	"github.com/tribehq/platform/models"
	"net/http"
)

var (
	// ErrInvalidUsernameOrPassword ...
	ErrInvalidUsernameOrPassword = errors.New("invalid username or password")
)

func passwordGrant(r *http.Request, client *models.OAuthApplication) (*AccessTokenResponse, error) {
	// Get the scope string
	scope, err := GetScope(r.Form.Get("scope"))
	if err != nil {
		return nil, err
	}

	// Authenticate the user
	user, err := AuthUser(r.Form.Get("username"), r.Form.Get("password"))
	if err != nil {
		// For security reasons, return a general error message
		return nil, ErrInvalidUsernameOrPassword
	}

	// Log in the user
	accessToken, refreshToken, err := Login(client, user, scope)
	if err != nil {
		return nil, err
	}

	// Create response
	accessTokenResponse, err := NewAccessTokenResponse(
		accessToken,
		refreshToken,
		config.AccessTokenLifetime,
		Bearer,
	)
	if err != nil {
		return nil, err
	}

	return accessTokenResponse, nil
}
