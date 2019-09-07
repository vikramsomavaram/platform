package oauth

import (
	"gitlab.com/mytribe/platform/controllers/oauth2/config"
	"gitlab.com/mytribe/platform/models"
	"net/http"
)

func clientCredentialsGrant(r *http.Request, client *models.OAuthApplication) (*AccessTokenResponse, error) {
	// Get the scope string
	scope, err := GetScope(r.Form.Get("scope"))
	if err != nil {
		return nil, err
	}

	// Create a new access token
	accessToken, err := GrantAccessToken(
		client,
		nil,                        // empty user
		config.AccessTokenLifetime, // expires in
		scope,
	)
	if err != nil {
		return nil, err
	}

	// Create response
	accessTokenResponse, err := NewAccessTokenResponse(
		accessToken,
		nil, // refresh token
		config.AccessTokenLifetime,
		Bearer,
	)
	if err != nil {
		return nil, err
	}

	return accessTokenResponse, nil
}
