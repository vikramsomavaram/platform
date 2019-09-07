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

func (r *queryResolver) Products(ctx context.Context, productType *models.ProductSearchType, text *string, productStatus *models.ProductStatus, after *string, before *string, first *int, last *int) (*models.ProductConnection, error) {
	var items []*models.Product
	var edges []*models.ProductEdge
	filter := bson.D{}
	limit := 25

	items, totalCount, hasPrevious, hasNext, err := models.GetProducts(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.ProductEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.ProductConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

func (r *queryResolver) Product(ctx context.Context, id primitive.ObjectID) (*models.Product, error) {
	product := models.GetProductByID(id.Hex())
	if product.ID.IsZero() {
		return nil, ErrProductNotFound
	}
	return product, nil
}

// productResolver is of type struct.
type productResolver struct{ *Resolver }

func (r *productResolver) Type(ctx context.Context, obj *models.Product) (string, error) {
	panic("implement me")
}

func (r *productResolver) GroupedProducts(ctx context.Context, obj *models.Product) ([]string, error) {
	panic("implement me")
}

func (r *productResolver) Variations(ctx context.Context, obj *models.Product) ([]string, error) {
	panic("implement me")
}

func (r *productResolver) Status(ctx context.Context, obj *models.Product) (models.ProductStatus, error) {
	panic("implement me")
}

func (r *productResolver) ServiceType(ctx context.Context, obj *models.Product) (models.StoreCategory, error) {
	panic("implement me")
}

func (r *mutationResolver) AddProduct(ctx context.Context, input models.AddProductInput) (*models.Product, error) {
	product := &models.Product{}
	_ = copier.Copy(&product, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	product.CreatedBy = user.ID
	product, err = models.CreateProduct(*product)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), product.ID.Hex(), "product", product, nil, ctx)
	return product, nil
}

func (r *mutationResolver) UpdateProduct(ctx context.Context, input models.UpdateProductInput) (*models.Product, error) {
	product := &models.Product{}
	product = models.GetProductByID(input.ID.Hex())
	if product.ID.IsZero() {
		return nil, ErrProductNotFound
	}
	_ = copier.Copy(&product, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	product.CreatedBy = user.ID
	product, err = models.UpdateProduct(product)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), product.ID.Hex(), "product", product, nil, ctx)
	return product, nil
}

func (r *mutationResolver) DeleteProduct(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteProductByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "product", nil, nil, ctx)
	return &res, err
}

func (r *mutationResolver) ActivateProduct(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	product := models.GetProductByID(id.Hex())
	if product.ID.IsZero() {
		return nil, ErrProductNotFound
	}
	product.IsActive = true
	_, err := models.UpdateProduct(product)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "product", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

func (r *mutationResolver) DeactivateProduct(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	product := models.GetProductByID(id.Hex())
	if product.ID.IsZero() {
		return nil, ErrProductNotFound
	}
	product.IsActive = false
	_, err := models.UpdateProduct(product)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "product", nil, nil, ctx)
	return utils.PointerBool(false), nil
}
