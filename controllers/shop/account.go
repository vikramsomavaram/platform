/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package shop

import (
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"net/http"
)

func GetAccountPage(ctx echo.Context) error {
	return ctx.Render(http.StatusOK, "cart", "")
}

func AccountLogin(ctx echo.Context) error {
	return ctx.Render(http.StatusOK, "cart", "")
}

func AccountLogout(ctx echo.Context) error {
	sess, _ := session.Get("session", ctx)
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}
	sess.Values["useremail"] = ""
	sess.Values["authenticated"] = ""
	_ = sess.Save(ctx.Request(), ctx.Response())
	return ctx.Redirect(http.StatusSeeOther, "/account/login")
	//return ctx.Render(http.StatusOK, "cart", "")
}

func GetAccountAddresses(ctx echo.Context) error {
	return ctx.Render(http.StatusOK, "cart", "")
}

func GetAccountLoginPage(ctx echo.Context) error {
	return ctx.Render(http.StatusOK, "customers.login", "")
}

func GetAccountOrders(ctx echo.Context) error {
	return ctx.Render(http.StatusOK, "cart", "")
}
