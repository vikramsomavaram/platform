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

//EmailTemplate returns an email template by its ID
func (r *queryResolver) EmailTemplate(ctx context.Context, templateID primitive.ObjectID) (*models.EmailTemplate, error) {
	emailTemplate, err := models.GetEmailTemplateByID(templateID.String())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return emailTemplate, nil

}

type emailTemplateResolver struct{ *Resolver }

//EmailTemplates gives a list of email templates
func (r *queryResolver) EmailTemplates(ctx context.Context, emailTemplateSearchType *models.EmailTemplateSearchType, text *string, after *string, before *string, first *int, last *int) (*models.EmailTemplateConnection, error) {
	var items []*models.EmailTemplate
	var edges []*models.EmailTemplateEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetEmailTemplates(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.EmailTemplateEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.EmailTemplateConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//GetEmailTemplateDetails returns email template details by their IDs
func (r *queryResolver) GetEmailTemplateDetails(ctx context.Context, templateID *string) (*models.EmailTemplate, error) {
	emailTemplateDetails, err := models.GetEmailTemplateByID(*templateID)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return emailTemplateDetails, nil
}

//AddEmailTemplate adds a new email template
func (r *mutationResolver) AddEmailTemplate(ctx context.Context, input models.AddEmailTemplateInput) (*models.EmailTemplate, error) {
	emailTemplate := &models.EmailTemplate{}
	_ = copier.Copy(&emailTemplate, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	emailTemplate.CreatedBy = user.ID
	emailTemplate, err = models.CreateEmailTemplate(emailTemplate)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), emailTemplate.ID.Hex(), "email template", emailTemplate, nil, ctx)
	return emailTemplate, nil
}

//UpdateEmailTemplate updates an existing email template
func (r *mutationResolver) UpdateEmailTemplate(ctx context.Context, input models.UpdateEmailTemplateInput) (*models.EmailTemplate, error) {
	emailTemplate := &models.EmailTemplate{}
	emailTemplate, err := models.GetEmailTemplateByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&emailTemplate, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	emailTemplate.CreatedBy = user.ID
	emailTemplate, err = models.UpdateEmailTemplate(emailTemplate)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), emailTemplate.ID.Hex(), "email template", emailTemplate, nil, ctx)
	return emailTemplate, nil
}

//DeleteEmailTemplate deletes an existing email template
func (r *mutationResolver) DeleteEmailTemplate(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteEmailTemplateByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "email template", nil, nil, ctx)
	return &res, err
}
