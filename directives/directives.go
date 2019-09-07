/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package directives

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/tribehq/platform/models"
	"github.com/tribehq/platform/resolvers"
	"github.com/tribehq/platform/utils/auth"
	"github.com/vektah/gqlparser/gqlerror"
	"strings"
)

// Directives ...
var Directives = resolvers.DirectiveRoot{
	IsAuthenticated: IsAuthenticated,
	HasScope:        HasScope,
}

// IsAuthenticated is used to authenticate.
func IsAuthenticated(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
	user, err := auth.ForContext(ctx)
	if user == nil || err != nil {
		err = &gqlerror.Error{Message: "invalid authentication", Extensions: map[string]interface{}{"code": "invalid_authentication"}}
		return nil, err
	}

	return next(ctx)
}

// HasScope defines scope.
func HasScope(ctx context.Context, obj interface{}, next graphql.Resolver, scopes []string) (interface{}, error) {
	user, err := auth.ForContext(ctx)
	//Check the scope on the user account role here
	if user == nil || err != nil {
		err = &gqlerror.Error{Message: "resource access forbidden", Extensions: map[string]interface{}{"code": "unauthorized_client"}}
		return nil, err
	}
	dbPerms := models.GetMergedUserPermissions(user.ID.Hex())
	serviceMatched := false
	permissionMatched := false
	for _, scope := range scopes {
		vars := strings.Split(scope, ":")
		scopeService, scopePermission := vars[0], vars[1]
		for _, dbPerm := range *dbPerms {
			dbvars := strings.Split(dbPerm, ":")
			dbService, dbPermission := dbvars[0], dbvars[1]
			if dbService == "*" || dbService == scopeService {
				serviceMatched = true
			}
			if dbPermission == "*" || dbPermission == scopePermission {
				permissionMatched = true
			}
		}
	}
	if serviceMatched && permissionMatched {
		return next(ctx)
	}

	err = &gqlerror.Error{Message: "resource access forbidden", Extensions: map[string]interface{}{"code": "unauthorized_client"}}
	return nil, err
}
