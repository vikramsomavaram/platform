package oauth

import (
	"github.com/tribehq/platform/controllers/oauth2/config"
	"github.com/tribehq/platform/models"
	"net/http"
)

func refreshTokenGrant(r *http.Request, client *models.OAuthApplication) (*AccessTokenResponse, error) {
	// Fetch the refresh token
	theRefreshToken, err := GetValidRefreshToken(r.Form.Get("refresh_token"), client)
	if err != nil {
		return nil, err
	}

	// Get the scope
	scope, err := getRefreshTokenScope(theRefreshToken, r.Form.Get("scope"))
	if err != nil {
		return nil, err
	}

	// Log in the user
	accessToken, refreshToken, err := Login(
		theRefreshToken.Client(),
		theRefreshToken.User(),
		scope,
	)
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
