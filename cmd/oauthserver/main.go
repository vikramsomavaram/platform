/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package main

import (
	"context"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/gorilla/sessions"
	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
	"github.com/tribehq/platform/controllers/oauth2/oauth"
	sessionservice "github.com/tribehq/platform/controllers/oauth2/session"
	"github.com/tribehq/platform/controllers/oauth2/web"
	"github.com/tribehq/platform/lib/cache"
	"github.com/tribehq/platform/lib/database"
	"github.com/tribehq/platform/lib/log/echo_logger"
	"github.com/tribehq/platform/lib/log/log_formatter"
	smw "github.com/tribehq/platform/middleware"
	"github.com/tribehq/platform/utils/echo_template"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.opencensus.io/trace"
	"html/template"
	"io"
	"net/http"
	"os"
)

func init() {
	//Logrus Settings
	log.SetFormatter(log_formatter.NewFormatter())
	log.SetReportCaller(true)
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// OAuth2 Server
func main() {
	// Create and register a OpenCensus Stackdriver Trace exporter.
	exporter, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: os.Getenv("GOOGLE_CLOUD_PROJECT"),
	})
	if err != nil {
		log.Fatal(err)
	}
	trace.RegisterExporter(exporter)

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	cookieStore := sessions.NewCookieStore([]byte(os.Getenv("COOKIE_SECRET")))
	e.Use(session.Middleware(cookieStore)) //Required for OAuth2 Shit

	t := echo_template.New(echo_template.TemplateConfig{
		Root:         "public/views",
		Extension:    ".html",
		Master:       "layouts/outside",
		DisableCache: true,
	})

	e.Renderer = t

	//Logging
	e.Use(echo_logger.LogrusLogger())

	//CORs
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	//GZip Stuff
	e.Use(middleware.Gzip())

	//JWT Auth
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(smw.AppEngineHeaders)

	//OAuth Related middlewares
	sessionService := sessionservice.NewService(cookieStore)
	guestMw := web.NewGuestMiddleware(*sessionService)
	loggedInMw := web.NewLoggedInMiddleware(*sessionService)
	clientMw := web.NewClientMiddleware(*sessionService)

	//Database & Cache
	database.ConnectMongo() //Connect to MongoDB
	cache.ConnectRedis()

	//OAuth2 Server Stuff
	e.GET("/", func(ctx echo.Context) error {
		return ctx.Redirect(http.StatusFound, "/login")
	})
	e.GET("/login", web.LoginForm, guestMw.Serve(), clientMw.Serve())
	e.POST("/login", web.Login, guestMw.Serve(), clientMw.Serve())
	e.GET("/signup", web.RegisterForm, guestMw.Serve())
	e.POST("/signup", web.Register, guestMw.Serve())
	e.GET("/logout", web.Logout, loggedInMw.Serve())
	e.GET("/authorize", web.AuthorizeForm, loggedInMw.Serve(), clientMw.Serve())
	e.POST("/authorize", web.Authorize, loggedInMw.Serve(), clientMw.Serve())
	e.POST("/v1/oauth/tokens", oauth.TokensHandler)
	e.POST("/v1/oauth/introspect", oauth.IntrospectHandler)

	//AppEngine Health Check
	e.Any("/_ah/health", func(ctx echo.Context) error {
		dbc := database.MongoDBClient
		err := dbc.Ping(context.Background(), &readpref.ReadPref{})
		if err != nil {
			log.Error(err)
			return ctx.String(http.StatusInternalServerError, "database error")
		}
		rc := cache.RedisClient
		err = rc.Ping().Err()
		if err != nil {
			log.Error(err)
			return ctx.String(http.StatusInternalServerError, "cache error")
		}
		return ctx.String(http.StatusOK, "ok")
	})

	e.Server.Addr = port("8080")
	log.Infoln("OAuth2 Server listening on: ", e.Server.Addr)
	e.Logger.Fatal(e.StartServer(e.Server)) //Listen on Given Port
}

func port(cfgPort string) string {
	var port = os.Getenv("PORT")
	if port == "" {
		port = cfgPort
	}
	return ":" + port
}
