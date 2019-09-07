/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package resolvers

import (
	"context"
	"encoding/base64"
	log "github.com/sirupsen/logrus"
	"github.com/tribehq/platform/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//NewsletterSubscribers gives a list of newsletter subscribers
func (r *queryResolver) NewsletterSubscribers(ctx context.Context, newsletterSubscriberType *models.NewsletterSubscriberType, newsletterSubscriberStatus *models.NewsletterSubscriberStatus, text *string, after *string, before *string, first *int, last *int) (*models.NewsletterSubscriberConnection, error) {
	var items []*models.NewsletterSubscriber
	var edges []*models.NewsletterSubscriberEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetNewsletterSubscribers(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.NewsletterSubscriberEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.NewsletterSubscriberConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//NewsletterSubscriber returns a newsletter subscriber
func (r *queryResolver) NewsletterSubscriber(ctx context.Context, id primitive.ObjectID) (*models.NewsletterSubscriber, error) {
	newsletterSubscriber, err := models.GetNewsletterSubscriberByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return newsletterSubscriber, nil
}
