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

func (r *queryResolver) ProductMetadatas(ctx context.Context, id primitive.ObjectID, after *string, before *string, first *int, last *int) (*models.ProductMetadataConnection, error) {
	var items []*models.ProductMetadata
	var edges []*models.ProductMetadataEdge
	filter := bson.D{}
	limit := 25

	items, totalCount, hasPrevious, hasNext, err := models.GetProductMetadatas(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.ProductMetadataEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	ProductMetadataList := &models.ProductMetadataConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return ProductMetadataList, nil
}

func (r *queryResolver) ProductMetadata(ctx context.Context, id primitive.ObjectID) (*models.ProductMetadata, error) {
	productMetadata := models.GetProductMetadataByID(id.Hex())
	return productMetadata, nil

}

func (r *mutationResolver) AddProductMetadata(ctx context.Context, input models.AddProductMetadataInput) (*models.ProductMetadata, error) {
	productMetadata := &models.ProductMetadata{}
	_ = copier.Copy(&productMetadata, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	productMetadata.CreatedBy = user.ID
	productMetadata, err = models.CreateProductMetadata(*productMetadata)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), productMetadata.ID.Hex(), "product image", productMetadata, nil, ctx)
	return productMetadata, nil
}

func (r *mutationResolver) UpdateProductMetadata(ctx context.Context, input models.UpdateProductMetadataInput) (*models.ProductMetadata, error) {
	productMetadata := &models.ProductMetadata{}
	productMetadata = models.GetProductMetadataByID(input.ID.Hex())
	_ = copier.Copy(&productMetadata, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	productMetadata.CreatedBy = user.ID
	productMetadata = models.UpdateProductMetadata(productMetadata)
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), productMetadata.ID.Hex(), "product image", productMetadata, nil, ctx)
	return productMetadata, nil
}

func (r *mutationResolver) DeleteProductMetadata(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteProductMetadataByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "product image", nil, nil, ctx)
	return &res, err
}
