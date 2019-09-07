/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package resolvers

import (
	"context"
	"encoding/base64"
	"github.com/jinzhu/copier"
	log "github.com/sirupsen/logrus"
	"github.com/tribehq/platform/lib/audit_log"
	"github.com/tribehq/platform/models"
	"github.com/tribehq/platform/utils/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//SmsTemplates gives a list of SMS templates
func (r *queryResolver) SmsTemplates(ctx context.Context, smsTemplateType *models.SMSTemplateSearchType, text *string, after *string, before *string, first *int, last *int) (*models.SMSTemplateConnection, error) {
	var items []*models.SMSTemplate
	var edges []*models.SMSTemplateEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetSMSTemplates(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.SMSTemplateEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.SMSTemplateConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

type smsTemplateResolver struct{ *Resolver }

func (smsTemplateResolver) Title(ctx context.Context, obj *models.SMSTemplate) (string, error) {
	return obj.Title, nil
}

//SmsTemplate returns an SMS template by its ID
func (r *queryResolver) SmsTemplate(ctx context.Context, templateID primitive.ObjectID) (*models.SMSTemplate, error) {
	smsTemplate, err := models.GetSMSTemplateByID(templateID.String())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return smsTemplate, nil
}

//AddSmsTemplate adds a new SMS template
func (r *mutationResolver) AddSmsTemplate(ctx context.Context, input models.AddSmsTemplateInput) (*models.SMSTemplate, error) {
	smsTemplate := &models.SMSTemplate{}
	_ = copier.Copy(&smsTemplate, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	smsTemplate.CreatedBy = user.ID
	smsTemplate, err = models.CreateSMSTemplate(smsTemplate)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), smsTemplate.ID.Hex(), "sms template", smsTemplate, nil, ctx)
	return smsTemplate, nil
}

//UpdateSmsTemplate updates an existing SMS template
func (r *mutationResolver) UpdateSmsTemplate(ctx context.Context, input models.UpdateSmsTemplateInput) (*models.SMSTemplate, error) {
	smsTemplate := &models.SMSTemplate{}
	smsTemplate, err := models.GetSMSTemplateByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&smsTemplate, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	smsTemplate.CreatedBy = user.ID
	smsTemplate, err = models.UpdateSMSTemplate(smsTemplate)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), smsTemplate.ID.Hex(), "sms template", smsTemplate, nil, ctx)
	return smsTemplate, nil
}

//DeleteSmsTemplate deletes an existing SMS template
func (r *mutationResolver) DeleteSmsTemplate(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteSMSTemplateByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "sms template", nil, nil, ctx)
	return &res, err
}

type sMSTemplateResolver struct{ *Resolver }
