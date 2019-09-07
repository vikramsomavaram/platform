/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package main

import (
	"context"
	"contrib.go.opencensus.io/exporter/stackdriver"
	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/osteele/liquid"
	log "github.com/sirupsen/logrus"
	"github.com/tribehq/platform/cmd/shopserver/tags"
	"github.com/tribehq/platform/controllers/shop"
	"github.com/tribehq/platform/lib/cache"
	"github.com/tribehq/platform/lib/database"
	"github.com/tribehq/platform/lib/log/echo_logger"
	"github.com/tribehq/platform/lib/log/log_formatter"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.opencensus.io/trace"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Template struct {
	themeRoot    string
	templatesMap map[string]string
	engine       *liquid.Engine
}

func (t *Template) Render(w io.Writer, name string, data interface{}, ctx echo.Context) error {
	var bindings liquid.Bindings
	switch data.(type) {
	case map[string]interface{}:
		bindings = data.(map[string]interface{})
	default:
		bindings = map[string]interface{}{}
	}
	if t.templatesMap[name] != "" {
		templateContent, err := ioutil.ReadFile(t.templatesMap[name])
		if err != nil {
			log.Error("error reading template file %q: %v\n", t.templatesMap[name], err)
		}
		parsedTemplate, terr := t.engine.ParseTemplateLocation(templateContent, t.templatesMap[name], 0)
		if terr != nil {
			log.Error(terr)
		}
		out, terr := parsedTemplate.Render(bindings)
		if terr != nil {
			log.Error(terr)
		}

		_, err = w.Write(out)
		return err
	}
	return ctx.String(http.StatusInternalServerError, "invalid template")
}

func (t *Template) parseTemplates() {
	templatesMap := make(map[string]string)
	//walks the theme root
	err := filepath.Walk(t.themeRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatalf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		//We only need *.liquid files to parse
		if filepath.Ext(path) == ".liquid" {
			templateName := strings.TrimSuffix(strings.TrimPrefix(path, "theme/"), ".liquid")
			templatesMap[templateName] = path
		}
		return nil
	})
	if err != nil {
		log.Fatalf("error walking the path %q: %v\n", t.themeRoot, err)
	}
	t.templatesMap = templatesMap
}

// Shopify Liquid rendering Server
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

	t := &Template{
		themeRoot: "./theme",
		engine:    liquid.NewEngine(),
	}
	//Add needed tags & blocks
	//Tags
	t.engine.RegisterTag("section", tags.SectionTag)
	t.engine.RegisterTag("layout", tags.LayoutTag)

	//Blocks
	t.engine.RegisterBlock("paginate", tags.StyleTag)
	t.engine.RegisterBlock("javascript", tags.StyleTag)
	t.engine.RegisterBlock("stylesheet", tags.StyleTag)
	t.engine.RegisterBlock("form", tags.FormTag)
	t.engine.RegisterBlock("schema", tags.SchemaTag)
	t.engine.RegisterBlock("style", tags.StyleTag)
	t.parseTemplates()
	e.Renderer = t

	//Logging
	e.Use(echo_logger.LogrusLogger())

	//Logrus Settings
	log.SetFormatter(log_formatter.NewFormatter())
	log.SetReportCaller(true)

	//CORs
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	//GZip Stuff
	e.Use(middleware.Gzip())

	//JWT Auth
	//e.Use(smw.JWT([]byte("jwtsecret")))
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())

	//e.Use(smw.AppEngineHeaders)
	database.ConnectMongo() //Connect to MongoDB
	cache.ConnectRedis()

	//TODO Middleware to get all the store settings and inject it
	//e.Use(shopifyMw.StoreDomainSettings)
	//Handles store session authentication
	//e.Use(shopifyMw.StoreSessionAuth)

	//Load and Process store theme details here

	////Handle store i18n
	//// I18nBundle ...
	//var I18nBundle *localization.Bundle
	//bundle := localization.NewBundle(language.English)
	//
	////Load all language bundles
	////TODO load up from database using AddMessages method instead of files
	//_, err = bundle.LoadMessageFile("es.toml")
	//if err != nil {
	//	log.Error()
	//}

	//Initialize RBAC
	//auth.InitRBAC()

	//Routes
	//Home

	e.GET("/", shop.GetHomePage)
	//Cart

	e.GET("/cart", shop.GetCartPage)
	//Account
	e.GET("/account", shop.GetAccountPage)

	//Account Login
	e.GET("/account/login", shop.GetAccountLoginPage)

	//Account Login
	e.POST("/account/login", shop.AccountLogin)

	//Account Login
	e.GET("/account/logout", shop.AccountLogout)

	//Account Addresses
	e.GET("/account/addresses", shop.GetAccountAddresses)

	//Account Orders
	e.GET("/account/orders", shop.GetAccountOrders)

	//Search
	e.GET("/search", shop.GetSearchPage)
	e.POST("/search", shop.GetSearchResultsPage)

	//Site blog posts
	e.GET("/blogs", shop.GetBlogHomePage)

	//Single site blog
	e.GET("/blogs/:slug", shop.GetBlogArticlePage)

	//Product collections
	e.GET("/collections", shop.GetProductCollectionsHomePage)

	//Single product collection
	e.GET("/collections/:slug", shop.GetSingleProductCollectionPage)

	//Products
	e.GET("/products", shop.GetProductsPage)
	e.GET("/products/:slug", shop.GetSingleProductPage)

	//Pages
	e.GET("/pages/:slug", shop.GetSitePage)

	//Checkout
	e.GET("/checkout/:cartId", shop.GetCheckoutCart)

	//TODO handle 404 and other edge cases
	//Sitemaps for products / collections / pages / blogs etc.,
	e.GET("/sitemap.xml", shop.RootSiteMap)
	e.GET("/sitemap_products.xml", shop.ProductsSiteMap)
	e.GET("/sitemap_collections.xml", shop.ProductCollectionsSiteMap)
	e.GET("/sitemap_pages.xml", shop.PagesSiteMap)
	e.GET("/sitemap_blogs.xml", shop.BlogSiteMap)
	e.GET("/opensearch_description.xml", shop.OpensearchDescription)

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
	log.Infoln("Shopify Server listening on: ", e.Server.Addr)
	e.Logger.Fatal(e.StartServer(e.Server))
}

func port(cfgPort string) string {
	var port = os.Getenv("PORT")
	if port == "" {
		port = cfgPort
	}
	return ":" + port
}
