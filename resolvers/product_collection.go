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

//AddProductCollection adds new product collection.
func (r *mutationResolver) AddProductCollection(ctx context.Context, input models.AddProductCollectionInput) (*models.ProductCollection, error) {
	productCollection := &models.ProductCollection{}
	_ = copier.Copy(&productCollection, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	productCollection.CreatedBy = user.ID
	productCollection, err = models.CreateProductCollection(*productCollection)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), productCollection.ID.Hex(), "product collection", productCollection, nil, ctx)
	return productCollection, nil
}

//UpdateProductCollection updates product Collection.
func (r *mutationResolver) UpdateProductCollection(ctx context.Context, input models.UpdateProductCollectionInput) (*models.ProductCollection, error) {
	productCollection := &models.ProductCollection{}
	productCollection = models.GetProductCollectionByID(input.ID.Hex())
	_ = copier.Copy(&productCollection, &input)

	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	productCollection.CreatedBy = user.ID
	productCollection, err = models.UpdateProductCollection(productCollection)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), productCollection.ID.Hex(), "product collection", productCollection, nil, ctx)
	return productCollection, nil
}

//DeleteProductCollection deletes product collection.
func (r *mutationResolver) DeleteProductCollection(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteProductCollectionByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "product collection", nil, nil, ctx)
	return &res, err
}

//DeactivateProductCollection deactivates product collection.
func (r *mutationResolver) DeactivateProductCollection(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	productCollection := models.GetProductCollectionByID(id.Hex())
	productCollection.IsActive = false
	_, err := models.UpdateProductCollection(productCollection)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "product collection", nil, nil, ctx)
	return utils.PointerBool(false), nil
}

//ActivateProductCollection activates product collection.
func (r *mutationResolver) ActivateProductCollection(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	productCollection := models.GetProductCollectionByID(id.Hex())
	productCollection.IsActive = true
	_, err := models.UpdateProductCollection(productCollection)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "product collection", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//ProductCollections returns a list of product collections.
func (r *queryResolver) ProductCollections(ctx context.Context, id primitive.ObjectID, after *string, before *string, first *int, last *int) (*models.ProductCollectionConnection, error) {
	var items []*models.ProductCollection
	var edges []*models.ProductCollectionEdge
	filter := bson.D{}
	limit := 25

	items, totalCount, hasPrevious, hasNext, err := models.GetProductCollections(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.ProductCollectionEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	productCollectionList := &models.ProductCollectionConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return productCollectionList, nil
}

//ProductCollection returns product collection by id.
func (r *queryResolver) ProductCollection(ctx context.Context, id primitive.ObjectID) (*models.ProductCollection, error) {
	productCollection := models.GetProductCollectionByID(id.Hex())
	return productCollection, nil
}
