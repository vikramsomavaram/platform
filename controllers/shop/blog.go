/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package shop

import (
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"github.com/tribehq/platform/models"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
)

func GetBlogHomePage(ctx echo.Context) error {
	articles, totalCount, hasPrevious, hasNext, err := models.GetBlogPosts(bson.D{}, 20, nil, nil, nil, nil)
	if err != nil {
		log.Error(err)
		return ctx.Render(http.StatusInternalServerError, "500", "")
	}
	log.Print(totalCount, hasPrevious, hasNext)
	return ctx.Render(http.StatusOK, "blog", echo.Map{"blog": map[string]interface{}{
		"articles": articles,
	}})
}

func GetBlogArticlePage(ctx echo.Context) error {
	return ctx.Render(http.StatusOK, "article", "")
}
