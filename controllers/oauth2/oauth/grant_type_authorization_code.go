package oauth

import (
	"errors"
	"github.com/prometheus/common/log"
	"github.com/tribehq/platform/controllers/oauth2/config"
	"github.com/tribehq/platform/models"
	"net/http"
)

var (
	// ErrInvalidRedirectURI ...
	ErrInvalidRedirectURI = errors.New("invalid redirect URI")
)

func authorizationCodeGrant(r *http.Request, client *models.OAuthApplication) (*AccessTokenResponse, error) {
	// Fetch the authorization code
	authorizationCode, err := getValidAuthorizationCode(
		r.Form.Get("code"),
		r.Form.Get("redirect_uri"),
		client,
	)
	if err != nil {
		return nil, err
	}

	// Log in the user
	accessToken, refreshToken, err := Login(
		authorizationCode.Client,
		authorizationCode.User,
		authorizationCode.Scope,
	)
	if err != nil {
		return nil, err
	}

	// Delete the authorization code
	_, err = models.DeleteAuthorizationCodeByID(authorizationCode.ID.Hex())
	if err != nil {
		log.Error(err)
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
