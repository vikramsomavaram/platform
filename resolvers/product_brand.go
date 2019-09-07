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

func (r *mutationResolver) AddProductBrand(ctx context.Context, input models.AddProductBrandInput) (*models.ProductBrand, error) {
	productBrand := &models.ProductBrand{}
	_ = copier.Copy(&productBrand, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	productBrand.CreatedBy = user.ID
	productBrand, err = models.CreateProductBrand(*productBrand)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), productBrand.ID.Hex(), "product brand", productBrand, nil, ctx)
	return productBrand, nil
}

func (r *mutationResolver) UpdateProductBrand(ctx context.Context, input models.UpdateProductBrandInput) (*models.ProductBrand, error) {
	productBrand := &models.ProductBrand{}
	productBrand = models.GetProductBrandByID(input.ID.Hex())
	_ = copier.Copy(&productBrand, &input)

	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	productBrand.CreatedBy = user.ID
	productBrand, err = models.UpdateProductBrand(productBrand)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), productBrand.ID.Hex(), "product brand", productBrand, nil, ctx)
	return productBrand, nil
}

func (r *mutationResolver) DeleteProductBrand(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteProductBrandByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "product brand", nil, nil, ctx)
	return &res, err
}

func (r *mutationResolver) DeactivateProductBrand(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	productBrand := models.GetProductBrandByID(id.Hex())
	productBrand.IsActive = false
	_, err := models.UpdateProductBrand(productBrand)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "product brand", nil, nil, ctx)
	return utils.PointerBool(false), nil
}

func (r *mutationResolver) ActivateProductBrand(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	productBrand := models.GetProductBrandByID(id.Hex())
	productBrand.IsActive = true
	_, err := models.UpdateProductBrand(productBrand)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "product brand", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

func (r *queryResolver) ProductBrands(ctx context.Context, brandID primitive.ObjectID, after *string, before *string, first *int, last *int) (*models.ProductBrandConnection, error) {
	var items []*models.ProductBrand
	var edges []*models.ProductBrandEdge
	filter := bson.D{}
	limit := 25

	items, totalCount, hasPrevious, hasNext, err := models.GetProductBrands(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.ProductBrandEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	productBrandList := &models.ProductBrandConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return productBrandList, nil
}

func (r *queryResolver) ProductBrand(ctx context.Context, id primitive.ObjectID) (*models.ProductBrand, error) {
	productBrand := models.GetProductBrandByID(id.Hex())
	return productBrand, nil
}
