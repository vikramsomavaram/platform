package oauth

import (
	"github.com/tribehq/platform/controllers/oauth2/config"
	"github.com/tribehq/platform/models"
)

// Login creates an access token and refresh token for a user (logs him/her in)
func Login(client *models.OAuthApplication, user *models.User, scope string) (*models.AccessToken, *models.RefreshToken, error) {
	// Return error if user's role is not allowed to use this service
	//if !IsRoleAllowed(user.Roles,[]string{"user"}) {
	//	// For security reasons, return a general error message
	//	return nil, nil, ErrInvalidUsernameOrPassword
	//}

	// Create a new access token
	accessToken, err := GrantAccessToken(
		client,
		user,
		config.AccessTokenLifetime, // expires in
		scope,
	)

	if err != nil {
		return nil, nil, err
	}

	// Create or retrieve a refresh token
	refreshToken, err := GetOrCreateRefreshToken(
		client,
		user,
		config.RefreshTokenLifetime, // expires in
		scope,
	)
	if err != nil {
		return nil, nil, err
	}

	return accessToken, refreshToken, nil
}
