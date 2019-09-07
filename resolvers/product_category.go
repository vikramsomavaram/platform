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

//AddProductCategory adds product category.
func (r *mutationResolver) AddProductCategory(ctx context.Context, input models.AddProductCategoryInput) (*models.ProductCategory, error) {
	productCategory := &models.ProductCategory{}
	_ = copier.Copy(&productCategory, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	productCategory.CreatedBy = user.ID
	productCategory, err = models.CreateProductCategory(*productCategory)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), productCategory.ID.Hex(), "product category", productCategory, nil, ctx)
	return productCategory, nil
}

//UpdateProductCategory updates product category.
func (r *mutationResolver) UpdateProductCategory(ctx context.Context, input models.UpdateProductCategoryInput) (*models.ProductCategory, error) {
	productCategory := &models.ProductCategory{}
	productCategory = models.GetProductCategoryByID(input.ID.Hex())
	_ = copier.Copy(&productCategory, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	productCategory.CreatedBy = user.ID
	productCategory, err = models.UpdateProductCategory(productCategory)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), productCategory.ID.Hex(), "product category", productCategory, nil, ctx)
	return productCategory, nil
}

//DeleteProductCategory deletes product category.
func (r *mutationResolver) DeleteProductCategory(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteProductCategoryByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "product category", nil, nil, ctx)
	return &res, err
}

//ActivateProductCategory activates product category.
func (r *mutationResolver) ActivateProductCategory(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	productCategory := models.GetProductCategoryByID(id.Hex())
	productCategory.IsActive = true
	_, err := models.UpdateProductCategory(productCategory)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "product category", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//DeactivateProductCategory deactivates product category.
func (r *mutationResolver) DeactivateProductCategory(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	productCategory := models.GetProductCategoryByID(id.Hex())
	productCategory.IsActive = false
	_, err := models.UpdateProductCategory(productCategory)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "product category", nil, nil, ctx)
	return utils.PointerBool(false), nil
}

//ProductCategories returns a list of product categories.
func (r *queryResolver) ProductCategories(ctx context.Context, productCategoryType *models.ItemCategoryType, text *string, productCategoryStatus *models.ProductStatus, after *string, before *string, first *int, last *int) (*models.ProductCategoryConnection, error) {
	var items []*models.ProductCategory
	var edges []*models.ProductCategoryEdge
	filter := bson.D{}
	limit := 25

	items, totalCount, hasPrevious, hasNext, err := models.GetProductCategories(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.ProductCategoryEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	productCategoryList := &models.ProductCategoryConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return productCategoryList, nil
}

//ProductCategory gives product category by id.
func (r *queryResolver) ProductCategory(ctx context.Context, id primitive.ObjectID) (*models.ProductCategory, error) {
	productCategory := models.GetProductCategoryByID(id.Hex())
	return productCategory, nil
}
