package oauth

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/tribehq/platform/models"
	"go.mongodb.org/mongo-driver/bson"
	"sort"
	"strings"
)

var (
	// ErrInvalidScope ...
	ErrInvalidScope = errors.New("invalid scope")
)

// GetScope takes a requested scope and, if it's empty, returns the default
// scope, if not empty, it validates the requested scope
func GetScope(requestedScope string) (string, error) {
	// Return the default scope if the requested scope is empty
	if requestedScope == "" {
		return GetDefaultScope(), nil
	}

	// If the requested scope exists in the database, return it
	if ScopeExists(requestedScope) {
		return requestedScope, nil
	}

	// Otherwise return error
	return "", ErrInvalidScope
}

// GetDefaultScope returns the default scope
func GetDefaultScope() string {
	// Fetch default scopes
	var scopes []string
	oauthScopes, _, _, _, err := models.GetOAuthScopesByFilter(bson.D{{"isDefault", true}}, 1000, nil, nil, nil, nil)
	if err != nil {
		log.Error(err)
	}

	for _, oauthScope := range oauthScopes {
		scopes = append(scopes, oauthScope.Scope)
	}

	// Sort the scopes alphabetically
	sort.Strings(scopes)

	// Return space delimited scope string
	return strings.Join(scopes, " ")
}

// ScopeExists checks if a scope exists
func ScopeExists(requestedScope string) bool {
	// Split the requested scope string
	scopes := strings.Split(requestedScope, " ")

	// Count how many of requested scopes exist in the database
	var count int64
	_, count, _, _, err := models.GetOAuthScopesByFilter(bson.D{{"scope", bson.M{"$in": scopes}}}, 1000, nil, nil, nil, nil)
	if err != nil {
		log.Error(err)
	}

	// Return true only if all requested scopes found
	return count == int64(len(scopes))
}
