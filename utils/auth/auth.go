/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package auth

import (
	"context"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/tribehq/platform/models"
)

var echoCtxKey = &contextKey{"echo"}

type contextKey struct {
	name string
}

//EchoContextToGraphQLContext make echo.Context available in resolver function context
func EchoContextToGraphQLContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		newCtx := context.WithValue(ctx.Request().Context(), echoCtxKey, ctx)
		ctx.SetRequest(ctx.Request().WithContext(newCtx))
		return next(ctx)
	}
}

//ForContext finds the user from the context.
func ForContext(ctx context.Context) (*models.User, error) {
	echoContext := ctx.Value(echoCtxKey).(echo.Context)
	if echoContext.Get("user") != nil {
		userToken := echoContext.Get("user").(*jwt.Token)
		claims := userToken.Claims.(jwt.MapClaims)
		uid := claims["id"].(string)
		user := models.GetUserByID(uid)
		return user, nil
	}

	return nil, errors.New("invalid user authentication")
}

func JwtToken(ctx context.Context) string {
	echoContext := ctx.Value(echoCtxKey).(echo.Context)
	if echoContext.Get("user") != nil {
		jwtToken := echoContext.Get("user").(*jwt.Token)
		return jwtToken.Raw
	}
	return ""
}
