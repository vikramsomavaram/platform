/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package middleware

import (
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

//AppEngineHeaders gets country, state and city from headers set by Google AppEngine
// https://cloud.google.com/appengine/docs/standard/go/reference/request-response-headers
func AppEngineHeaders(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		country := ctx.Request().Header.Get("X-AppEngine-Country")
		city := ctx.Request().Header.Get("X-AppEngine-City")
		state := ctx.Request().Header.Get("X-AppEngine-Region")
		userIP := ctx.Request().Header.Get("X-AppEngine-User-IP")
		latLng := ctx.Request().Header.Get("X-AppEngine-CityLatLong")
		log.Println("User request originated from: Country:" + country + "-> State/Region: " + state + "-> City: " + city + "-> IP Address: " + userIP + "->" + latLng)
		return next(ctx)
	}
}
