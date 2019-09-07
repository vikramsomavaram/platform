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

//AddDeliveryCharges adds a new delivery charge
func (r *mutationResolver) AddDeliveryCharges(ctx context.Context, input models.AddDeliveryChargeInput) (*models.DeliveryCharge, error) {
	deliveryCharge := &models.DeliveryCharge{}
	_ = copier.Copy(&deliveryCharge, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	deliveryCharge.CreatedBy = user.ID
	deliveryCharge, err = models.CreateDeliveryCharge(*deliveryCharge)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), deliveryCharge.ID.Hex(), "delivery charge", deliveryCharge, nil, ctx)
	return deliveryCharge, nil
}

//UpdateDeliveryCharges updates an existing delivery charge
func (r *mutationResolver) UpdateDeliveryCharges(ctx context.Context, input models.UpdateDeliveryChargeInput) (*models.DeliveryCharge, error) {
	deliveryCharge := &models.DeliveryCharge{}
	deliveryCharge, err := models.GetDeliveryChargeByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&deliveryCharge, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	deliveryCharge.CreatedBy = user.ID
	deliveryCharge, err = models.UpdateDeliveryCharge(deliveryCharge)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), deliveryCharge.ID.Hex(), "delivery charge", deliveryCharge, nil, ctx)
	return deliveryCharge, nil
}

//DeleteDeliveryCharges deletes an existing delivery charge
func (r *mutationResolver) DeleteDeliveryCharges(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteDeliveryChargeByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "delivery charge", nil, nil, ctx)
	return &res, err
}

//DeactivateDeliveryCharges deactivates a delivery charge by its ID
func (r *mutationResolver) DeactivateDeliveryCharges(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	deliveryCharge, err := models.GetDeliveryChargeByID(id.Hex())
	if err != nil {
		return nil, err
	}
	deliveryCharge.IsActive = false
	_, err = models.UpdateDeliveryCharge(deliveryCharge)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "delivery charge", nil, nil, ctx)
	return utils.PointerBool(false), nil

}

//ActivateDeliveryCharges activates a delivery charge by its ID
func (r *mutationResolver) ActivateDeliveryCharges(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	deliveryCharge, err := models.GetDeliveryChargeByID(id.Hex())
	if err != nil {
		return nil, err
	}
	deliveryCharge.IsActive = true

	_, err = models.UpdateDeliveryCharge(deliveryCharge)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "delivery charge", nil, nil, ctx)
	return utils.PointerBool(true), nil

}

//DeliveryCharges gives a list of delivery charges
func (r *queryResolver) DeliveryCharges(ctx context.Context, deliveryChargeSearch *models.DeliveryChargesSearch, text *string, after *string, before *string, first *int, last *int) (*models.DeliveryChargeConnection, error) {
	var items []*models.DeliveryCharge
	var edges []*models.DeliveryChargeEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetDeliveryCharges(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.DeliveryChargeEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.DeliveryChargeConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//DeliveryCharge returns a delivery charge by its ID
func (r *queryResolver) DeliveryCharge(ctx context.Context, deliveryChargeID primitive.ObjectID) (*models.DeliveryCharge, error) {
	company, err := models.GetDeliveryChargeByID(deliveryChargeID.String())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return company, nil

}

type deliveryChargeResolver struct{ *Resolver }
