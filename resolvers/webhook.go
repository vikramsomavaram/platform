/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package resolvers

import (
	"context"
	"encoding/base64"
	"errors"
	"github.com/jinzhu/copier"
	log "github.com/sirupsen/logrus"
	"github.com/tribehq/platform/lib/audit_log"
	"github.com/tribehq/platform/models"
	"github.com/tribehq/platform/utils"
	"github.com/tribehq/platform/utils/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

//AddWebhook adds a new webhook
func (r *mutationResolver) AddWebhook(ctx context.Context, appID primitive.ObjectID, input models.AddWebhookInput) (*models.Webhook, error) {
	if appID.IsZero() {
		return nil, errors.New("appID is required")
	}
	webhook := &models.Webhook{}
	_ = copier.Copy(&webhook, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	webhook.CreatedBy = user.ID
	//TODO check if the appId exists
	webhook.AppID = appID.String()
	webhook, err = models.CreateWebhook(*webhook)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), webhook.ID.Hex(), "webhook", webhook, nil, ctx)
	return webhook, nil
}

//UpdateWebhook updates an existing webhook
func (r *mutationResolver) UpdateWebhook(ctx context.Context, input models.UpdateWebhookInput) (*models.Webhook, error) {
	webhook := &models.Webhook{}
	webhook, err := models.GetWebhookByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&webhook, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	webhook.CreatedBy = user.ID
	webhook.UpdatedAt = time.Now()
	webhook, err = models.UpdateWebhook(webhook)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), webhook.ID.Hex(), "webhook", webhook, nil, ctx)
	return webhook, nil

}

//DeleteWebhook deletes an existing webhook
func (r *mutationResolver) DeleteWebhook(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteWebhookByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "webhook", nil, nil, ctx)
	return &res, err
}

//ActivateWebhook activates a webhook by its ID
func (r *mutationResolver) ActivateWebhook(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	webhook, err := models.GetWebhookByID(id.Hex())
	if err != nil {
		return nil, err
	}
	webhook.IsActive = true
	_, err = models.UpdateWebhook(webhook)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "webhook", nil, nil, ctx)
	return utils.PointerBool(true), nil

}

//DeactivateWebhook deactivates a webhook by its ID
func (r *mutationResolver) DeactivateWebhook(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	webhook, err := models.GetWebhookByID(id.Hex())
	if err != nil {
		return nil, err
	}
	webhook.IsActive = false
	_, err = models.UpdateWebhook(webhook)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "webhook", nil, nil, ctx)
	return utils.PointerBool(false), nil

}

//Webhooks give a list of webhooks
func (r *queryResolver) Webhooks(ctx context.Context, appID primitive.ObjectID, after *string, before *string, first *int, last *int) (*models.WebhooksConnection, error) {
	var items []*models.Webhook
	var edges []*models.WebhookEdge
	filter := bson.D{}
	filter = append(filter, bson.E{"appId", appID})
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetWebhooks(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.WebhookEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.WebhooksConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//Webhook returns a webhook by ID
func (r *queryResolver) Webhook(ctx context.Context, id primitive.ObjectID) (*models.Webhook, error) {
	webhook, err := models.GetWebhookByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return webhook, nil
}

//TODO : check this later
//WebhookStatistics gives a list of webhook statistics
func (r *queryResolver) WebhookStatistics(ctx context.Context, id primitive.ObjectID) (*models.WebhookStatistics, error) {
	panic("implement me")
}

//WebhookLogs gives a list of webhook logs
func (r *queryResolver) WebhookLogs(ctx context.Context, id primitive.ObjectID, after *string, before *string, first *int, last *int) (*models.WebhookLogConnection, error) {
	var items []*models.WebhookLog
	var edges []*models.WebhookLogEdge
	filter := bson.D{}
	filter = append(filter, bson.E{"webhookId", id})
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetWebhookLogs(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.WebhookLogEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.WebhookLogConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//WebHookLog returns a webhook log by its ID
func (r *queryResolver) WebHookLog(ctx context.Context, id primitive.ObjectID) (*models.WebhookLog, error) {
	webHookLog, err := models.GetWebhookLogByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return webHookLog, nil
}

// webhookResolver is of type struct.
type webhookResolver struct{ *Resolver }

//IsEnabled
func (r *webhookResolver) IsEnabled(ctx context.Context, obj *models.Webhook) (bool, error) {
	return obj.IsActive, nil
}

// webhookLogResolver is of type struct.
type webhookLogResolver struct{ *Resolver }

// CreatedAt
func (r *webhookLogResolver) CreatedAt(ctx context.Context, obj *models.WebhookLog) (string, error) {
	return obj.CreatedAt.String(), nil
}
