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
	"github.com/tribehq/platform/utils"
	"github.com/tribehq/platform/utils/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//DeliveryChargesUtilities gives a list of delivery charges utility
func (r *queryResolver) DeliveryChargesUtilities(ctx context.Context, deliveryChargesUtilityType *models.DeliveryChargesUtilityType, text *string, deliveryChargesUtilityStatus *models.DeliveryChargesUtilityStatus, after *string, before *string, first *int, last *int) (*models.DeliveryChargesUtilityConnection, error) {
	var items []*models.DeliveryChargesUtility
	var edges []*models.DeliveryChargesUtilityEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetDeliveryChargesUtilities(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.DeliveryChargesUtilityEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.DeliveryChargesUtilityConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//DeliveryChargesUtility returns a delivery charge utility by its ID
func (r *queryResolver) DeliveryChargesUtility(ctx context.Context, id primitive.ObjectID) (*models.DeliveryChargesUtility, error) {
	deliveryChargesUtility, err := models.GetDeliveryChargesUtilityByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return deliveryChargesUtility, nil
}

//AddDeliveryChargesUtility adds a new delivery charge utility
func (r *mutationResolver) AddDeliveryChargesUtility(ctx context.Context, input models.AddDeliveryChargesUtilityInput) (*models.DeliveryChargesUtility, error) {
	deliveryChargesUtility := &models.DeliveryChargesUtility{}
	_ = copier.Copy(&deliveryChargesUtility, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	deliveryChargesUtility.CreatedBy = user.ID
	deliveryChargesUtility, err = models.CreateDeliveryChargesUtility(*deliveryChargesUtility)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), deliveryChargesUtility.ID.Hex(), "delivery charge utility", deliveryChargesUtility, nil, ctx)
	return deliveryChargesUtility, nil
}

//UpdateDeliveryChargesUtility updates an existing delivery charge utility
func (r *mutationResolver) UpdateDeliveryChargesUtility(ctx context.Context, input models.UpdateDeliveryChargesUtilityInput) (*models.DeliveryChargesUtility, error) {
	deliveryChargesUtility := &models.DeliveryChargesUtility{}
	deliveryChargesUtility, err := models.GetDeliveryChargesUtilityByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&deliveryChargesUtility, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	deliveryChargesUtility.CreatedBy = user.ID
	deliveryChargesUtility, err = models.UpdateDeliveryChargesUtility(deliveryChargesUtility)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), deliveryChargesUtility.ID.Hex(), "delivery charge utility", deliveryChargesUtility, nil, ctx)
	return deliveryChargesUtility, nil
}

//DeleteDeliveryChargesUtility deletes an existing delivery charge utility
func (r *mutationResolver) DeleteDeliveryChargesUtility(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteDeliveryChargesUtilityByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "delivery charge utility", nil, nil, ctx)
	return &res, err
}

//ActivateDeliveryChargesUtility activates a delivery charge utility by its ID
func (r *mutationResolver) ActivateDeliveryChargesUtility(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	deliveryChargesUtility, err := models.GetDeliveryChargesUtilityByID(id.Hex())
	if err != nil {
		return nil, err
	}
	deliveryChargesUtility.IsActive = true
	_, err = models.UpdateDeliveryChargesUtility(deliveryChargesUtility)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "delivery charge utility", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//DeactivateDeliveryChargesUtility deactivates a delivery charge utility by its ID
func (r *mutationResolver) DeactivateDeliveryChargesUtility(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	deliveryChargesUtility, err := models.GetDeliveryChargesUtilityByID(id.Hex())
	if err != nil {
		return nil, err
	}
	deliveryChargesUtility.IsActive = false
	_, err = models.UpdateDeliveryChargesUtility(deliveryChargesUtility)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "delivery charge utility", nil, nil, ctx)
	return utils.PointerBool(false), nil
}

// deliveryChargesUtilityResolver is of type struct.
type deliveryChargesUtilityResolver struct{ *Resolver }
