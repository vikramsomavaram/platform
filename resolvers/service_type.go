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

type serviceTypeResolver struct{ *Resolver }

func (r *serviceTypeResolver) ServiceCharge(ctx context.Context, obj *models.ServiceType) (float64, error) {
	return obj.ServiceCharge, nil
}

func (r *serviceTypeResolver) Commission(ctx context.Context, obj *models.ServiceType) (float64, error) {
	return obj.Commission, nil
}

//ServiceTypes gives a list of service types
func (r *queryResolver) ServiceTypes(ctx context.Context, text *string, serviceTypeSubCategory *string, serviceTypeStatus *models.ServiceTypeStatus, after *string, before *string, first *int, last *int) (*models.ServiceTypeConnection, error) {
	var items []*models.ServiceType
	var edges []*models.ServiceTypeEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetServiceTypes(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.ServiceTypeEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.ServiceTypeConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//ServiceType returns a service type by ID
func (r *queryResolver) ServiceType(ctx context.Context, id primitive.ObjectID) (*models.ServiceType, error) {
	serviceType, err := models.GetServiceTypeByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return serviceType, nil
}

//AddServiceType adds a new service type
func (r *mutationResolver) AddServiceType(ctx context.Context, input models.AddServiceTypeInput) (*models.ServiceType, error) {
	serviceType := &models.ServiceType{}
	_ = copier.Copy(&serviceType, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	serviceType.CreatedBy = user.ID
	serviceType, err = models.CreateServiceType(*serviceType)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), serviceType.ID.Hex(), "service type", serviceType, nil, ctx)
	return serviceType, nil
}

//UpdateServiceType updates an existing service type
func (r *mutationResolver) UpdateServiceType(ctx context.Context, input models.UpdateServiceTypeInput) (*models.ServiceType, error) {
	serviceType := &models.ServiceType{}
	serviceType, err := models.GetServiceTypeByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&serviceType, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	serviceType.CreatedBy = user.ID
	serviceType, err = models.UpdateServiceType(serviceType)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), serviceType.ID.Hex(), "service type", serviceType, nil, ctx)
	return serviceType, nil
}

//DeleteServiceType deletes an existing service type
func (r *mutationResolver) DeleteServiceType(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteServiceTypeByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "service type", nil, nil, ctx)
	return &res, err
}

//ActivateServiceType activates a service type by its ID
func (r *mutationResolver) ActivateServiceType(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	serviceType, err := models.GetServiceTypeByID(id.Hex())
	if err != nil {
		return nil, err
	}
	serviceType.IsActive = true
	_, err = models.UpdateServiceType(serviceType)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "service type", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//DeactivateServiceType deactivates a service type by its ID
func (r *mutationResolver) DeactivateServiceType(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	serviceType, err := models.GetServiceTypeByID(id.Hex())
	if err != nil {
		return nil, err
	}
	serviceType.IsActive = false
	_, err = models.UpdateServiceType(serviceType)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "service type", nil, nil, ctx)
	return utils.PointerBool(false), nil
}
