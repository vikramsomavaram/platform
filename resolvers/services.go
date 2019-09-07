/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package resolvers

import (
	"context"
	"encoding/base64"
	"github.com/jinzhu/copier"
	"github.com/tribehq/platform/lib/audit_log"
	"github.com/tribehq/platform/models"
	"github.com/tribehq/platform/utils"
	"github.com/tribehq/platform/utils/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//Services gives a list of services
func (r *queryResolver) Services(ctx context.Context, serviceStatus *models.ServiceStatus, after *string, before *string, first *int, last *int) (*models.ServiceConnection, error) {
	var items []*models.Service
	var edges []*models.ServiceEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetServices(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.ServiceEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.ServiceConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//Service returns a service by its ID
func (r *queryResolver) Service(ctx context.Context, id primitive.ObjectID) (*models.Service, error) {
	service := models.GetServiceByID(id.Hex())
	if service == nil || service.ID.IsZero() {
		return nil, ErrServiceNotFound
	}
	return service, nil
}

type serviceResolver struct{ *Resolver }

func (serviceResolver) SubCategories(ctx context.Context, obj *models.Service) ([]*models.ServiceSubCategory, error) {
	filter := bson.D{{"serviceId", obj.ID.Hex()}}
	subCategories, _, _, _, err := models.GetServiceSubCategories(filter, 10, nil, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	return subCategories, nil
}

//AddService adds a new service
func (r *mutationResolver) AddService(ctx context.Context, input models.AddServiceInput) (*models.Service, error) {
	service := &models.Service{}
	_ = copier.Copy(&service, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	service.CreatedBy = user.ID
	service, err = models.CreateService(*service)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), service.ID.Hex(), "service", service, nil, ctx)
	return service, nil
}

//UpdateService updates a new service
func (r *mutationResolver) UpdateService(ctx context.Context, input models.UpdateServiceInput) (*models.Service, error) {
	service := &models.Service{}
	service = models.GetServiceByID(input.ID.Hex())
	_ = copier.Copy(&service, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	service.CreatedBy = user.ID
	service, err = models.UpdateService(service)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), service.ID.Hex(), "service", service, nil, ctx)
	return service, nil
}

//DeleteService deletes an existing service
func (r *mutationResolver) DeleteService(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteServiceByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "service", nil, nil, ctx)
	return &res, err
	//TODO:delete respective sub categories
}

//ActivateService activates a service by its ID
func (r *mutationResolver) ActivateService(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	service := models.GetServiceByID(id.Hex())
	service.IsActive = true
	_, err := models.UpdateService(service)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "service", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//DeactivateService deactivate service by its ID
func (r *mutationResolver) DeactivateService(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	service := models.GetServiceByID(id.Hex())
	service.IsActive = false
	_, err := models.UpdateService(service)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "service", nil, nil, ctx)
	return utils.PointerBool(false), nil
}
