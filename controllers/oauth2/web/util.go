package web

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"net/url"
)

// Redirects to a new path while keeping current request's query string
func redirectWithQueryString(to string, query url.Values, ctx echo.Context) error {
	return ctx.Redirect(http.StatusFound, fmt.Sprintf("%s%s", to, getQueryString(query)))
}

// Redirects to a new path with the query string moved to the URL fragment
func redirectWithFragment(to string, ctx echo.Context) error {
	return ctx.Redirect(http.StatusFound, fmt.Sprintf("%s#%s", to, ctx.Request().URL.Query().Encode()))
}

// Returns string encoded query string of the request
func getQueryString(query url.Values) string {
	encoded := query.Encode()
	if len(encoded) > 0 {
		encoded = fmt.Sprintf("?%s", encoded)
	}
	return encoded
}

// Helper function to handle redirecting failed or declined authorization
func errorRedirect(ctx echo.Context, redirectURI *url.URL, err, state, responseType string) error {
	query := redirectURI.Query()
	query.Set("error", err)
	if state != "" {
		query.Set("state", state)
	}
	if responseType == "code" {
		return redirectWithQueryString(redirectURI.String(), query, ctx)
	}
	if responseType == "token" {
		return redirectWithFragment(redirectURI.String(), ctx)
	}
	//TODO check this later
	return redirectWithFragment(redirectURI.String(), ctx)
}
