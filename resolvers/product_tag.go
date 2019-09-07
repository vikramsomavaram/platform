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

func (r *queryResolver) ProductTags(ctx context.Context, id primitive.ObjectID, after *string, before *string, first *int, last *int) (*models.ProductTagConnection, error) {
	var items []*models.ProductTag
	var edges []*models.ProductTagEdge
	filter := bson.D{}
	limit := 25

	items, totalCount, hasPrevious, hasNext, err := models.GetProductTags(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.ProductTagEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	ProductTagList := &models.ProductTagConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return ProductTagList, nil
}

func (r *queryResolver) ProductTag(ctx context.Context, id primitive.ObjectID) (*models.ProductTag, error) {
	productTag := models.GetProductTagByID(id.Hex())
	return productTag, nil
}

func (r *mutationResolver) AddProductTag(ctx context.Context, input models.AddProductTagInput) (*models.ProductTag, error) {
	productTag := &models.ProductTag{}
	_ = copier.Copy(&productTag, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	productTag.CreatedBy = user.ID
	productTag, err = models.CreateProductTag(*productTag)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), productTag.ID.Hex(), "product tag", productTag, nil, ctx)
	return productTag, nil
}

func (r *mutationResolver) UpdateProductTag(ctx context.Context, input models.UpdateProductTagInput) (*models.ProductTag, error) {
	productTag := &models.ProductTag{}
	productTag = models.GetProductTagByID(input.ID.Hex())
	_ = copier.Copy(&productTag, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	productTag.CreatedBy = user.ID
	productTag = models.UpdateProductTag(productTag)
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), productTag.ID.Hex(), "product tag", productTag, nil, ctx)
	return productTag, nil
}

func (r *mutationResolver) DeleteProductTag(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteProductTagByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "product tag", nil, nil, ctx)
	return &res, err
}
