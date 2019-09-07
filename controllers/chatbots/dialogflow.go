/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package chatbots

import (
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"google.golang.org/genproto/googleapis/cloud/dialogflow/v2"
	"net/http"
)

// DialogFlowWebHookHandler entire fulfillment happens here for Voice Agents & Chat Bots
func DialogFlowWebHookHandler(ctx echo.Context) error {
	log.Debug("DialogFlow Stuff")

	var err error

	wr := dialogflow.WebhookRequest{}

	if err = ctx.Bind(&wr); err != nil {
		log.WithError(err).Error("Couldn't Unmarshal request to json")
		return ctx.JSON(http.StatusBadRequest, "")
	}

	contexts := wr.GetQueryResult().GetOutputContexts()
	for _, context := range contexts {
		log.Debugln(context.Name)
	}

	return ctx.JSON(http.StatusOK, "")
}
