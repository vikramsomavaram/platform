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
	"github.com/tribehq/platform/utils/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *queryResolver) ProductAttributes(ctx context.Context, id primitive.ObjectID, after *string, before *string, first *int, last *int) (*models.ProductAttributeConnection, error) {
	var items []*models.ProductAttribute
	var edges []*models.ProductAttributeEdge
	filter := bson.D{}
	limit := 25

	items, totalCount, hasPrevious, hasNext, err := models.GetProductAttributes(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.ProductAttributeEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	ProductAttributeList := &models.ProductAttributeConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return ProductAttributeList, nil
}

func (r *queryResolver) ProductAttribute(ctx context.Context, id primitive.ObjectID) (*models.ProductAttribute, error) {
	productAttribute, err := models.GetProductAttributeByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return productAttribute, nil
}

func (r *mutationResolver) AddProductAttribute(ctx context.Context, input models.AddProductAttributeInput) (*models.ProductAttribute, error) {
	productAttribute := &models.ProductAttribute{}
	_ = copier.Copy(&productAttribute, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	productAttribute.CreatedBy = user.ID
	productAttribute, err = models.CreateProductAttribute(*productAttribute)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), productAttribute.ID.Hex(), "product attribute", productAttribute, nil, ctx)
	return productAttribute, nil
}

func (r *mutationResolver) UpdateProductAttribute(ctx context.Context, input models.UpdateProductAttributeInput) (*models.ProductAttribute, error) {
	productAttribute := &models.ProductAttribute{}
	productAttribute, err := models.GetProductAttributeByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&productAttribute, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	productAttribute.CreatedBy = user.ID
	productAttribute = models.UpdateProductAttribute(productAttribute)
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), productAttribute.ID.Hex(), "product attribute", productAttribute, nil, ctx)
	return productAttribute, nil
}

func (r *mutationResolver) DeleteProductAttribute(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteProductAttributeByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "product attribute", nil, nil, ctx)
	return &res, err
}
