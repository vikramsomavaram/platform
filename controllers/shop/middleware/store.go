/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/tribehq/platform/models"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
)

func StoreDomainSettings(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		hostname := ctx.Request().Host
		store, err := models.GetStoreByFilter(bson.D{{"store_domain", hostname}})
		if err != nil {
			return ctx.Render(http.StatusInternalServerError, "internal_error", "")
		}
		if store != nil && store.ID.IsZero() {
			//there is no store by that name
			return ctx.Render(http.StatusNotFound, "store_not_found", "")
		}
		ctx.Set("store", store)
		return next(ctx)
	}
}

func StoreSessionAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		hostname := ctx.Request().Host
		store, err := models.GetStoreByFilter(bson.D{{"store_domain", hostname}})
		if err != nil {
			return ctx.Render(http.StatusInternalServerError, "internal_error", "")
		}
		if store != nil && store.ID.IsZero() {
			//there is no store by that name
			return ctx.Render(http.StatusInternalServerError, "store_not_found", "")
		}
		ctx.Set("store", store)
		return next(ctx)
	}
}
