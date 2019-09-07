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

var ErrServiceNotFound = errors.New("invalid service")

//ServiceSubCategories gives a list of service sub categories
func (r *queryResolver) ServiceSubCategories(ctx context.Context, text *string, serviceSubCategoryStatus *models.ServiceSubCategoryStatus, after *string, before *string, first *int, last *int) (*models.ServiceSubCategoryConnection, error) {
	var items []*models.ServiceSubCategory
	var edges []*models.ServiceSubCategoryEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetServiceSubCategories(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.ServiceSubCategoryEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.ServiceSubCategoryConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//ServiceSubCategory returns a service sub category by its ID
func (r *queryResolver) ServiceSubCategory(ctx context.Context, id primitive.ObjectID) (*models.ServiceSubCategory, error) {
	subCategory := models.GetServiceSubCategoryByID(id.Hex())
	return subCategory, nil
}

//AddServiceSubCategory adds a new service sub category
func (r *mutationResolver) AddServiceSubCategory(ctx context.Context, input models.AddServiceSubCategoryInput) (*models.ServiceSubCategory, error) {
	subCategory := &models.ServiceSubCategory{}
	_ = copier.Copy(&subCategory, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	service := models.GetServiceByID(input.ServiceID)
	if service == nil || service.ID.IsZero() {
		return nil, ErrServiceNotFound
	}
	subCategory.CreatedBy = user.ID
	subCategory, err = models.CreateServiceSubCategory(*subCategory)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), subCategory.ID.Hex(), "sub category", subCategory, nil, ctx)
	return subCategory, nil
}

//UpdateServiceSubCategory updates an existing service sub category
func (r *mutationResolver) UpdateServiceSubCategory(ctx context.Context, input models.UpdateServiceSubCategoryInput) (*models.ServiceSubCategory, error) {
	subCategory := &models.ServiceSubCategory{}
	subCategory = models.GetServiceSubCategoryByID(input.ID.Hex())
	_ = copier.Copy(&subCategory, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	service := models.GetServiceByID(input.ServiceID)
	if service == nil || service.ID.IsZero() {
		return nil, ErrServiceNotFound
	}
	subCategory.CreatedBy = user.ID
	subCategory, err = models.UpdateServiceSubCategory(subCategory)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), subCategory.ID.Hex(), "sub category", subCategory, nil, ctx)
	return subCategory, nil
}

//DeleteServiceSubCategory deletes an existing service sub category
func (r *mutationResolver) DeleteServiceSubCategory(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteServiceSubCategoryByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "sub category", nil, nil, ctx)
	return &res, err
}

//DeactivateServiceSubCategory deactivates a service sub category by its ID
func (r *mutationResolver) DeactivateServiceSubCategory(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	subCategory := models.GetServiceSubCategoryByID(id.Hex())
	if subCategory.ID.IsZero() {
		return utils.PointerBool(false), ErrServiceSubCategoryNotFound
	}
	subCategory.IsActive = false
	_, err := models.UpdateServiceSubCategory(subCategory)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "sub category", nil, nil, ctx)
	return utils.PointerBool(false), nil
}

//ActivateServiceSubCategory activates a service sub category by its ID
func (r *mutationResolver) ActivateServiceSubCategory(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	subCategory := models.GetServiceSubCategoryByID(id.Hex())
	if subCategory.ID.IsZero() {
		return utils.PointerBool(false), ErrServiceSubCategoryNotFound
	}
	subCategory.IsActive = true
	_, err := models.UpdateServiceSubCategory(subCategory)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "sub category", nil, nil, ctx)
	return utils.PointerBool(true), nil
}
