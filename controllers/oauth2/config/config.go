/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package config

const (
	AccessTokenLifetime  int = 3600
	RefreshTokenLifetime int = 15552000 //6 months
	AuthCodeLifetime     int = 600      //10Minutes

	//Session related
	SessionCookieSecret = "no cookies for you !@#@#"
	SessionCookiePath   = ""
	// MaxAge=0 means no 'Max-Age' attribute specified.
	// MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'.
	// MaxAge>0 means Max-Age attribute present and given in seconds.
	SessionCookieMaxAge int = 3600000
	// When you tag a cookie with the HttpOnly flag, it tells the browser that
	// this particular cookie should only be accessed by the server.
	// Any attempt to access the cookie from client script is strictly forbidden.
	SessionCookieHTTPOnly bool = true
)
