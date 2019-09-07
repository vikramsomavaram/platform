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

//ServiceVehicleTypes returns a list of service vehicle types
func (r *queryResolver) ServiceVehicleTypes(ctx context.Context, serviceVehicleType *models.ServiceVehicleServiceType, serviceVehicleTypeStatus *models.ServiceVehicleTypeStatus, text *string, after *string, before *string, first *int, last *int) (*models.ServiceVehicleTypeConnection, error) {
	var items []*models.ServiceVehicleType
	var edges []*models.ServiceVehicleTypeEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetServiceVehicleTypes(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.ServiceVehicleTypeEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.ServiceVehicleTypeConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

type serviceVehicleTypeResolver struct{ *Resolver }

//ServiceVehicleType returns a service vehicle typ by its ID
func (r *queryResolver) ServiceVehicleType(ctx context.Context, id primitive.ObjectID) (*models.ServiceVehicleType, error) {
	serviceVehicleType, err := models.GetServiceVehicleTypeByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return serviceVehicleType, nil
}

//UpdateServiceVehicleType updates an existing service vehicle type
func (r *mutationResolver) UpdateServiceVehicleType(ctx context.Context, input models.UpdateServiceVehicleTypeInput) (*models.ServiceVehicleType, error) {
	serviceVehicleType := &models.ServiceVehicleType{}
	serviceVehicleType, err := models.GetServiceVehicleTypeByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&serviceVehicleType, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	serviceVehicleType.CreatedBy = user.ID
	serviceVehicleType, err = models.UpdateServiceVehicleType(serviceVehicleType)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), serviceVehicleType.ID.Hex(), "service vehicle type", serviceVehicleType, nil, ctx)
	return serviceVehicleType, nil
}

//AddServiceVehicleType adds a new service vehicle type
func (r *mutationResolver) AddServiceVehicleType(ctx context.Context, input models.AddServiceVehicleTypeInput) (*models.ServiceVehicleType, error) {
	serviceVehicleType := &models.ServiceVehicleType{}
	_ = copier.Copy(&serviceVehicleType, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	serviceVehicleType.CreatedBy = user.ID
	serviceVehicleType, err = models.CreateServiceVehicleType(*serviceVehicleType)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), serviceVehicleType.ID.Hex(), "service vehicle type", serviceVehicleType, nil, ctx)
	return serviceVehicleType, nil
}

//ActivateServiceVehicleType activates a service vehicle type by ID
func (r *mutationResolver) ActivateServiceVehicleType(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	serviceVehicleType, err := models.GetServiceVehicleTypeByID(id.Hex())
	if err != nil {
		return nil, err
	}
	serviceVehicleType.IsActive = true
	_, err = models.UpdateServiceVehicleType(serviceVehicleType)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "service vehicle type", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//DeactivateServiceVehicleType deactivates a service vehicle type by ID
func (r *mutationResolver) DeactivateServiceVehicleType(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	serviceVehicleType, err := models.GetServiceVehicleTypeByID(id.Hex())
	if err != nil {
		return nil, err
	}
	serviceVehicleType.IsActive = false
	_, err = models.UpdateServiceVehicleType(serviceVehicleType)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "service vehicle type", nil, nil, ctx)
	return utils.PointerBool(false), nil
}

//DeleteServiceVehicleType deletes an existing service vehicle type
func (r *mutationResolver) DeleteServiceVehicleType(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteServiceVehicleTypeByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "service vehicle type", nil, nil, ctx)
	return &res, err
}
