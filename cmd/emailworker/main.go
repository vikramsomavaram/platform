/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package emailworker

import (
	"context"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"gitlab.com/mytribe/platform/lib/database"
	"gitlab.com/mytribe/platform/lib/log/log_formatter"
	"gitlab.com/mytribe/platform/models"
)

// PubSubMessage is the payload of a Pub/Sub event.
type PubSubMessage struct {
	Data []byte `json:"data"`
}

func init() {
	//Logrus Settings
	log.SetFormatter(log_formatter.NewFormatter())
	log.SetReportCaller(true)
	database.ConnectMongo() //Connect to MongoDB
}

// DeliverEmail consumes a Pub/Sub message and sends email.
func DeliverEmail(ctx context.Context, m PubSubMessage) error {
	message := &models.EmailMessage{}

	err := json.Unmarshal(m.Data, &message)
	if err != nil {
		log.Errorln(err)
		return err
	}

	err = models.SendEmail(message.From, message.To, message.TemplateID, message.Language, message.Data, message.Attachments)
	if err != nil {
		log.Errorln(err)
		return err
	}

	return nil
}
