/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package shop

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func RootSiteMap(ctx echo.Context) error {
	return ctx.XML(http.StatusOK, "Hi Auth!")
}

func ProductsSiteMap(ctx echo.Context) error {
	return ctx.XML(http.StatusOK, "Hi Auth!")
}

func ProductCollectionsSiteMap(ctx echo.Context) error {
	return ctx.XML(http.StatusOK, "Hi Auth!")
}

func BlogSiteMap(ctx echo.Context) error {
	return ctx.XML(http.StatusOK, "Hi Auth!")
}

func PagesSiteMap(ctx echo.Context) error {
	return ctx.XML(http.StatusOK, "Hi Auth!")
}

func OpensearchDescription(ctx echo.Context) error {
	return ctx.XML(http.StatusOK, "Hi Auth!")
}
