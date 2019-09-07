/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package pubsub

import (
	"cloud.google.com/go/pubsub"
	"context"
	log "github.com/sirupsen/logrus"
	"os"
)

// PubsubClient represents pubsub client.
var PubsubClient *pubsub.Client

func init() {
	pubsubClient, err := configurePubsub(os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if err != nil {
		log.Fatal(err)
	}

	PubsubClient = pubsubClient
}

func configurePubsub(projectID string) (*pubsub.Client, error) {

	PubsubTopicID := os.Getenv("PUBSUB_TOPIC")
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Create the topic if it doesn't exist.
	if exists, err := client.Topic(PubsubTopicID).Exists(ctx); err != nil {
		return nil, err
	} else if !exists {
		if _, err := client.CreateTopic(ctx, PubsubTopicID); err != nil {
			return nil, err
		}
	}
	return client, nil
}
