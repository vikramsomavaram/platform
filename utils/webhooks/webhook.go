/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package webhooks

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	ps "github.com/tribehq/platform/lib/pubsub"
	"github.com/tribehq/platform/utils"
	"time"
)

type WebhookEvent struct {
	ID        string      `json:"id"`
	CreatedAt time.Time   `json:"createdAt"`
	EventType string      `json:"type"`
	Data      interface{} `json:"data"`
}

func NewWebhookEvent(eventType string, data interface{}) {
	webhookEvent := &WebhookEvent{ID: utils.RandomIDGen(20), CreatedAt: time.Now(), EventType: eventType, Data: data}
	topic := ps.PubsubClient.Topic("webhooks_delivery")
	b, err := json.Marshal(webhookEvent)
	if err != nil {
		return
	}
	ctx := context.Background()
	_, err = topic.Publish(ctx, &pubsub.Message{Data: b}).Get(ctx)
	if err != nil {
		log.Errorln(err)
	}
}
