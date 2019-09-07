/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package shop

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func GetSitePage(ctx echo.Context) error {
	return ctx.Render(http.StatusOK, "blog_home", "")
}
