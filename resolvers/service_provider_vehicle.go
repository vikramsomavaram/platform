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

//ServiceProviderVehicles gives a list of service provider vehicles
func (r *queryResolver) ServiceProviderVehicles(ctx context.Context, vehicleType *models.ProviderVehicleType, vehicleStatus *models.ProviderVehicleStatus, text *string, after *string, before *string, first *int, last *int) (*models.ServiceProviderVehicleDetailsConnection, error) {
	var items []*models.ServiceProviderVehicleDetails
	var edges []*models.ServiceProviderVehicleDetailsEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetServiceProviderVehicles(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.ServiceProviderVehicleDetailsEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.ServiceProviderVehicleDetailsConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//ServiceProviderVehicle returns a service provider vehicle by ID
func (r *queryResolver) ServiceProviderVehicle(ctx context.Context, id primitive.ObjectID) (*models.ServiceProviderVehicleDetails, error) {
	serviceProviderVehicle, err := models.GetServiceProviderVehicleByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return serviceProviderVehicle, nil
}

//UpdateServiceProviderVehicle updates a service provider vehicle
func (r *mutationResolver) UpdateServiceProviderVehicle(ctx context.Context, input models.UpdateServiceProviderVehicleInput) (*models.ServiceProviderVehicleDetails, error) {
	serviceProviderVehicle := &models.ServiceProviderVehicleDetails{}
	serviceProviderVehicle, err := models.GetServiceProviderVehicleByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&serviceProviderVehicle, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	serviceProviderVehicle.CreatedBy = user.ID
	serviceProviderVehicle, err = models.UpdateServiceProviderVehicle(serviceProviderVehicle)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), serviceProviderVehicle.ID.Hex(), "service provider vehicle", serviceProviderVehicle, nil, ctx)
	return serviceProviderVehicle, nil
}

//AddServiceProviderVehicle adds a new service provider vehicle
func (r *mutationResolver) AddServiceProviderVehicle(ctx context.Context, input models.AddServiceProviderVehicleInput) (*models.ServiceProviderVehicleDetails, error) {
	providerVehicle := &models.ServiceProviderVehicleDetails{}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	providerVehicle.CreatedBy = user.ID
	_ = copier.Copy(&providerVehicle, &input)
	providerVehicle, err = models.CreateServiceProviderVehicle(*providerVehicle)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), providerVehicle.ID.Hex(), "service provider vehicle", providerVehicle, nil, ctx)
	return providerVehicle, nil
}

//ActivateServiceProviderVehicle activates a service provider vehicle by ID
func (r *mutationResolver) ActivateServiceProviderVehicle(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	serviceProviderVehicle, err := models.GetServiceProviderVehicleByID(id.Hex())
	if err != nil {
		return nil, err
	}
	serviceProviderVehicle.IsActive = true
	_, err = models.UpdateServiceProviderVehicle(serviceProviderVehicle)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "service provider vehicle", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//DeactivateServiceProviderVehicle deactivates a service provider vehicle by ID
func (r *mutationResolver) DeactivateServiceProviderVehicle(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	serviceProviderVehicle, err := models.GetServiceProviderVehicleByID(id.Hex())
	if err != nil {
		return nil, err
	}
	serviceProviderVehicle.IsActive = false
	_, err = models.UpdateServiceProviderVehicle(serviceProviderVehicle)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "service provider vehicle", nil, nil, ctx)
	return utils.PointerBool(false), nil
}

//DeleteServiceProviderVehicle deletes a service provider vehicle
func (r *mutationResolver) DeleteServiceProviderVehicle(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteServiceProviderVehicleByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "service provider vehicle", nil, nil, ctx)
	return &res, err
}

// serviceProviderVehicleDetailsResolver is of type struct.
type serviceProviderVehicleDetailsResolver struct{ *Resolver }
