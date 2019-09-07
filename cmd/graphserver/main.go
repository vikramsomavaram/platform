/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package main

import (
	"cloud.google.com/go/profiler"
	"context"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/99designs/gqlgen-contrib/gqlopencensus"
	"github.com/99designs/gqlgen/handler"
	"github.com/gorilla/websocket"
	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
	"github.com/tribehq/platform/controllers/chatbots"
	"github.com/tribehq/platform/controllers/payments"
	"github.com/tribehq/platform/directives"
	"github.com/tribehq/platform/lib/cache"
	"github.com/tribehq/platform/lib/database"
	"github.com/tribehq/platform/lib/log/echo_logger"
	"github.com/tribehq/platform/lib/log/log_formatter"
	smw "github.com/tribehq/platform/middleware"
	"github.com/tribehq/platform/resolvers"
	"github.com/tribehq/platform/utils/auth"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.opencensus.io/trace"
	"net/http"
	"os"
	"time"
)

func init() {
	//Logrus Settings
	log.SetFormatter(log_formatter.NewFormatter())
	log.SetReportCaller(true)
}

// GraphQL Based API Server
func main() {

	//Profiler initialization, best done as early as possible.
	if err := profiler.Start(profiler.Config{
		ProjectID: os.Getenv("GOOGLE_CLOUD_PROJECT"),
	}); err != nil {
		log.Fatal(err)
	}

	// Create and register a OpenCensus Stackdriver Trace exporter.
	exporter, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: os.Getenv("GOOGLE_CLOUD_PROJECT"),
	})
	if err != nil {
		log.Fatal(err)
	}
	trace.RegisterExporter(exporter)
	trace.AlwaysSample()

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	//Logging
	e.Use(echo_logger.LogrusLogger())

	//CORs
	e.Use(middleware.CORS())

	//GZip Stuff
	e.Use(middleware.Gzip())
	//JWT Auth
	e.Use(smw.JWT([]byte(os.Getenv("JWT_SECRET"))))
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(auth.EchoContextToGraphQLContext) //Make echo.Context available in resolver function context
	e.Use(smw.AppEngineHeaders)

	//GraphQL File upload settings
	var mb int64 = 1 << 20
	uploadMaxMemory := handler.UploadMaxMemory(32 * mb)
	uploadMaxSize := handler.UploadMaxSize(50 * mb)

	database.ConnectMongo() //Connect to MongoDB
	cache.ConnectRedis()

	//create apq cache
	apqCache, err := cache.NewAPQCache(cache.RedisClient, 24*time.Hour)
	if err != nil {
		log.Fatalf("cannot create APQ redis cache: %v", err)
	}

	//Initialize RBAC
	auth.InitRBAC()
	e.Any("/", echo.WrapHandler(handler.Playground("GraphQL Playground", "/graphql")))
	e.Any("/graphql", echo.WrapHandler(handler.GraphQL(
		resolvers.NewExecutableSchema(resolvers.Config{Resolvers: &resolvers.Resolver{}, Directives: directives.Directives}),
		handler.EnablePersistedQueryCache(apqCache),
		handler.Tracer(gqlopencensus.New()),
		handler.WebsocketUpgrader(websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}), uploadMaxMemory, uploadMaxSize)))

	hooks := e.Group("/hooks")
	//Stripe Payments Handling
	hooks.POST("/hooks/stripe", payments.StripeWebHookHandler)
	//Razorpay Payments Handling
	hooks.POST("/hooks/razorpay", payments.RazorPayWebHookHandler)
	//Paytm Payments Handling
	hooks.POST("/hooks/paytm", payments.PaytmWebHookHandler)
	//Braintree Payments Payments Handling
	hooks.POST("/hooks/braintree", payments.BraintreeWebHookHandler)
	//Wechat Payments Handling
	hooks.POST("/hooks/wechat", payments.WechatPayWebHookHandler)
	//AliPay Payments Handling
	hooks.POST("/hooks/alipay", payments.AliPayWebHookHandler)
	//Google DialogFlow - Google Actions, Cortana, Siri, Alexa & Chatbots
	hooks.POST("/hooks/dialogflow", chatbots.DialogFlowWebHookHandler)

	//Health Check
	e.Any("/health", func(ctx echo.Context) error {
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
	e.Logger.Fatal(e.StartServer(e.Server)) //Listen on Given Port
}

func port(cfgPort string) string {
	var port = os.Getenv("PORT")
	if port == "" {
		port = cfgPort
	}
	log.Infoln("Graph API Server listening on: ", port)
	return ":" + port
}
