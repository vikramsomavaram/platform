/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package payments

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

// AliPayWebHookHandler handles razor pay webhooks.
func AliPayWebHookHandler(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, "")
}
