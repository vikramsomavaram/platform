package oauth

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/tribehq/platform/controllers/oauth2/config"
	"github.com/tribehq/platform/controllers/oauth2/session"
	"github.com/tribehq/platform/models"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

var (
	// ErrAccessTokenNotFound ...
	ErrAccessTokenNotFound = errors.New("access token not found")
	// ErrAccessTokenExpired ...
	ErrAccessTokenExpired = errors.New("access token expired")
)

// Authenticate checks the access token is valid
func Authenticate(token string) (*models.AccessToken, error) {
	// Fetch the access token from the database
	accessToken := new(models.AccessToken)
	accessToken, err := models.GetAccessTokenByFilter(bson.D{{"accessToken", token}})
	if err != nil {
		log.Error(err)
		return nil, errors.New("internal server error")
	}
	if accessToken == nil && err == nil {
		return nil, ErrAccessTokenNotFound
	}

	// Check the access token hasn't expired
	if time.Now().UTC().After(accessToken.ExpiresAt) {
		return nil, ErrAccessTokenExpired
	}

	// Extend refresh token expiration database
	filter := bson.D{{"clientId", accessToken.ClientID}}

	if accessToken.UserID != "" {
		filter = append(filter, bson.E{"userId", accessToken.UserID})
	}
	//else {
	//	query = query.Where("user_id IS NULL")
	//}
	refreshToken := models.GetRefreshTokenByFilter(filter)

	increasedExpiresAt := refreshToken.ExpiresAt.Add(time.Duration(config.RefreshTokenLifetime) * time.Second)

	refreshToken.ExpiresAt = increasedExpiresAt

	refreshToken, err = models.UpdateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	return accessToken, nil
}

// ClearUserTokens deletes the user's access and refresh tokens associated with this client id
func ClearUserTokens(userSession *session.UserSession) {
	// Clear all refresh tokens with user_id and client_id
	refreshToken := models.GetRefreshTokenByFilter(bson.D{{"token", userSession.RefreshToken}})
	if !refreshToken.ID.IsZero() {
		_, err := models.DeleteRefreshTokensByFilter(bson.D{{"clientId", refreshToken.ClientID}, {"userId", refreshToken.UserID}})
		if err != nil {
			log.Error(err)
		}
	}

	// Clear all access tokens with user_id and client_id
	accessToken, err := models.GetAccessTokenByFilter(bson.D{{"accessToken", userSession.AccessToken}})
	if err != nil {
		log.Error(err)
	}
	if !accessToken.ID.IsZero() {
		_, err := models.DeleteRefreshTokensByFilter(bson.D{{"clientId", refreshToken.ClientID}, {"userId", refreshToken.UserID}})
		if err != nil {
			log.Error(err)
		}
	}
}
