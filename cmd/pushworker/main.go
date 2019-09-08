/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package pushworker

import (
	"context"
	"encoding/json"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	log "github.com/sirupsen/logrus"
	"github.com/tribehq/platform/lib/database"
	"github.com/tribehq/platform/lib/log/log_formatter"
)

//This thing does the push notifications thing

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

// PushNotification represents push notification.
type PushNotification struct {
	FCMToken  string            `json:"fcm_token"`
	Title     string            `json:"title"`
	Body      string            `json:"body"`
	Data      map[string]string `json:"data"`
	Topic     string            `json:"topic"`
	Condition string            `json:"condition"`
}

// DeliverPush consumes a Pub/Sub message and delivers
// push notification via firebase cloud messaging.
func DeliverPush(ctx context.Context, m PubSubMessage) error {

	req := &PushNotification{}
	err := json.Unmarshal(m.Data, &req)
	if err != nil {
		log.Errorln(err)
		return err
	}

	//firebase app setup
	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		log.Fatalf("error initializing firebase app: %v\n", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		log.Fatalf("error getting firebase Messaging client: %v\n", err)
	}

	message := &messaging.Message{}

	//TODO handle all edge cases mentioned here: https://github.com/firebase/firebase-admin-go/blob/be821cdc340e4dcaf1abf04910f727c8b8bda086/snippets/messaging.go
	if req.Topic != "" {
		message.Data = req.Data
		message.Topic = req.Topic
	}

	if req.FCMToken != "" && req.Topic == "" {
		message.Data = req.Data
		message.Token = req.FCMToken
	}

	if req.Condition != "" {
		message.Data = req.Data
		message.Token = req.FCMToken
		message.Condition = req.Condition
	}

	// Send a message to the device corresponding to the provided
	// registration token.
	response, err := client.Send(ctx, message)
	if err != nil {
		log.Fatalln(err)
		return err
	}
	// Response is a message ID string.
	log.Println("Successfully sent message:", response)

	return nil
}
