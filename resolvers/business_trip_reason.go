/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package resolvers

import (
	"context"
	"encoding/base64"
	"errors"
	"github.com/jinzhu/copier"
	"github.com/tribehq/platform/lib/audit_log"
	"github.com/tribehq/platform/models"
	"github.com/tribehq/platform/utils"
	"github.com/tribehq/platform/utils/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//BusinessTripReasons gives a list of business trip reasons
func (r *queryResolver) BusinessTripReasons(ctx context.Context, reasonType *models.BusinessTripReasonType, text *string, after *string, before *string, first *int, last *int) (*models.BusinessTripReasonConnection, error) {
	var items []*models.BusinessTripReason
	var edges []*models.BusinessTripReasonEdge
	filter := bson.D{}
	limit := 25

	items, totalCount, hasPrevious, hasNext, err := models.GetBusinessTripReasons(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.BusinessTripReasonEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.BusinessTripReasonConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//BusinessTripReason returns a given business trip reason by its ID
func (r *queryResolver) BusinessTripReason(ctx context.Context, id primitive.ObjectID) (*models.BusinessTripReason, error) {
	businessTripReason := models.GetBusinessTripReasonByID(id.Hex())
	return businessTripReason, nil
}

//AddBusinessTripReason adds a new business trip reason
func (r *mutationResolver) AddBusinessTripReason(ctx context.Context, input models.AddBusinessTripReasonInput) (*models.BusinessTripReason, error) {
	businessTripReason := &models.BusinessTripReason{}
	_ = copier.Copy(&businessTripReason, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	businessTripReason.CreatedBy = user.ID
	businessTripReason, err = models.CreateBusinessTripReason(*businessTripReason)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), businessTripReason.ID.Hex(), "business trip reason", businessTripReason, nil, ctx)
	return businessTripReason, nil
}

//UpdateBusinessTripReason updates an existing business trip reason
func (r *mutationResolver) UpdateBusinessTripReason(ctx context.Context, input models.UpdateBusinessTripReasonInput) (*models.BusinessTripReason, error) {
	businessTripReason := &models.BusinessTripReason{}
	businessTripReason = models.GetBusinessTripReasonByID(input.ID.Hex())
	_ = copier.Copy(&businessTripReason, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	businessTripReason.CreatedBy = user.ID
	businessTripReason, err = models.UpdateBusinessTripReason(businessTripReason)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), businessTripReason.ID.Hex(), "business trip reason", businessTripReason, nil, ctx)
	return businessTripReason, nil
}

//DeleteBusinessTripReason deletes an existing business trip reason
func (r *mutationResolver) DeleteBusinessTripReason(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteBusinessTripReasonByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "business trip reason", nil, nil, ctx)
	return &res, err
}

//ActivateBusinessTripReason activates a business trip reason by its ID
func (r *mutationResolver) ActivateBusinessTripReason(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	businessTripReason := models.GetBusinessTripReasonByID(id.Hex())
	if businessTripReason.ID.IsZero() {
		return utils.PointerBool(false), errors.New("business trip reason not found")
	}
	businessTripReason.IsActive = true
	_, err := models.UpdateBusinessTripReason(businessTripReason)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "business trip reason", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//DeactivateBusinessTripReason deactivates a business trip reason by its ID
func (r *mutationResolver) DeactivateBusinessTripReason(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	businessTripReason := models.GetBusinessTripReasonByID(id.Hex())
	if businessTripReason.ID.IsZero() {
		return utils.PointerBool(false), errors.New("business trip reason not found")
	}
	businessTripReason.IsActive = false
	_, err := models.UpdateBusinessTripReason(businessTripReason)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "business trip reason", nil, nil, ctx)
	return utils.PointerBool(false), nil

}

type businessTripReasonResolver struct{ *Resolver }
