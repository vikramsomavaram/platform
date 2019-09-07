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

//DeliveryVehicleTypes gives a list of delivery vehicle types
func (r *queryResolver) DeliveryVehicleTypes(ctx context.Context, deliveryVehicleType *models.DeliveryVehicleSearchType, text *string, deliveryVehicleTypeStatus *models.DeliveryVehicleTypeStatus, after *string, before *string, first *int, last *int) (*models.DeliveryVehicleTypeConnection, error) {
	var items []*models.DeliveryVehicleType
	var edges []*models.DeliveryVehicleTypeEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetDeliveryVehicleTypes(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.DeliveryVehicleTypeEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.DeliveryVehicleTypeConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

// DeliveryVehicleType returns a specific delivery vehicle type by its ID
func (r *queryResolver) DeliveryVehicleType(ctx context.Context, id primitive.ObjectID) (*models.DeliveryVehicleType, error) {
	vehicleType, err := models.GetDeliveryVehicleTypeByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return vehicleType, nil
}

//AddDeliveryVehicleType adds a new delivery vehicle type
func (r *mutationResolver) AddDeliveryVehicleType(ctx context.Context, input models.AddDeliveryVehicleTypeInput) (*models.DeliveryVehicleType, error) {
	vehicleType := &models.DeliveryVehicleType{}
	_ = copier.Copy(&vehicleType, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	vehicleType.CreatedBy = user.ID
	vehicleType, err = models.CreateDeliveryVehicleType(*vehicleType)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), vehicleType.ID.Hex(), "delivery vehicle type", vehicleType, nil, ctx)
	return vehicleType, nil
}

//UpdateDeliveryVehicleType updates an existing delivery vehicle type
func (r *mutationResolver) UpdateDeliveryVehicleType(ctx context.Context, input models.UpdateDeliveryVehicleTypeInput) (*models.DeliveryVehicleType, error) {
	vehicleType := &models.DeliveryVehicleType{}
	vehicleType, err := models.GetDeliveryVehicleTypeByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&vehicleType, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	vehicleType.CreatedBy = user.ID
	vehicleType, err = models.UpdateDeliveryVehicleType(vehicleType)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), vehicleType.ID.Hex(), "delivery vehicle type", vehicleType, nil, ctx)
	return vehicleType, nil
}

// DeleteDeliveryVehicleType deletes an existing delivery vehicle type
func (r *mutationResolver) DeleteDeliveryVehicleType(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteDeliveryVehicleTypeByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "delivery vehicle type", nil, nil, ctx)
	return &res, err
}

//ActivateDeliveryVehicleType activates a delivery vehicle type by its ID
func (r *mutationResolver) ActivateDeliveryVehicleType(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	vehicleType, err := models.GetDeliveryVehicleTypeByID(id.Hex())
	if err != nil {
		return nil, err
	}
	vehicleType.IsActive = true
	_, err = models.UpdateDeliveryVehicleType(vehicleType)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "delivery vehicle type", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

////DeactivateDeliveryVehicleType deactivates a delivery vehicle type by its ID
func (r *mutationResolver) DeactivateDeliveryVehicleType(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	vehicleType, err := models.GetDeliveryVehicleTypeByID(id.Hex())
	if err != nil {
		return nil, err
	}
	vehicleType.IsActive = false
	_, err = models.UpdateDeliveryVehicleType(vehicleType)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "delivery vehicle type", nil, nil, ctx)
	return utils.PointerBool(false), nil
}

// deliveryVehicleTypeResolver is of type struct.
type deliveryVehicleTypeResolver struct{ *Resolver }
