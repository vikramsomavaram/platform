package oauth

import (
	"github.com/tribehq/platform/models"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

// GrantAccessToken deletes old tokens and grants a new access token
func GrantAccessToken(client *models.OAuthApplication, user *models.User, expiresIn int, scope string) (*models.AccessToken, error) {
	_, err := models.DeleteAccessTokenByFilter(bson.D{{"clientId", client.ID.Hex()}, {"userId", user.ID.Hex()}, {"expiresAt", bson.M{"$lte": time.Now()}}})
	if err != nil {
		return nil, err
	}

	// Create a new access token
	accessToken := models.NewOAuthAccessToken(client, user, expiresIn, scope)
	accessToken, err = models.CreateAccessToken(accessToken)
	if err != nil {
		return nil, err
	}

	return accessToken, nil
}
