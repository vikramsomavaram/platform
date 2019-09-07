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
	"github.com/tribehq/platform/utils/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *queryResolver) ProductImages(ctx context.Context, id primitive.ObjectID, after *string, before *string, first *int, last *int) (*models.ProductImageConnection, error) {
	var items []*models.ProductImage
	var edges []*models.ProductImageEdge
	filter := bson.D{}
	limit := 25

	items, totalCount, hasPrevious, hasNext, err := models.GetProductImages(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.ProductImageEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	ProductImageList := &models.ProductImageConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return ProductImageList, nil
}

func (r *queryResolver) ProductImage(ctx context.Context, id primitive.ObjectID) (*models.ProductImage, error) {
	productImage := models.GetProductImageByID(id.Hex())
	return productImage, nil
}

func (r *mutationResolver) AddProductImage(ctx context.Context, input models.AddProductImageInput) (*models.ProductImage, error) {
	productImage := &models.ProductImage{}
	_ = copier.Copy(&productImage, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	productImage.CreatedBy = user.ID
	productImage, err = models.CreateProductImage(*productImage)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), productImage.ID.Hex(), "product image", productImage, nil, ctx)
	return productImage, nil
}

func (r *mutationResolver) UpdateProductImage(ctx context.Context, input models.UpdateProductImageInput) (*models.ProductImage, error) {
	productImage := &models.ProductImage{}
	productImage = models.GetProductImageByID(input.ID.Hex())
	_ = copier.Copy(&productImage, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	productImage.CreatedBy = user.ID
	productImage = models.UpdateProductImage(productImage)
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), productImage.ID.Hex(), "product image", productImage, nil, ctx)
	return productImage, nil
}

func (r *mutationResolver) DeleteProductImage(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteProductImageByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "product image", nil, nil, ctx)
	return &res, err
}
