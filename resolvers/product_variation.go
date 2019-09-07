/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package resolvers

import (
	"context"
	"encoding/base64"
	"errors"
	"github.com/jinzhu/copier"
	log "github.com/sirupsen/logrus"
	"github.com/tribehq/platform/lib/audit_log"
	"github.com/tribehq/platform/models"
	"github.com/tribehq/platform/utils/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ErrParentProductNotFound = errors.New("invalid parent product")

func (r *mutationResolver) AddProductVariation(ctx context.Context, input models.AddProductVariationInput) (*models.ProductVariation, error) {
	productVariation := &models.ProductVariation{}
	_ = copier.Copy(&productVariation, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	product := models.GetProductByID(input.ParentProductID)
	if product == nil || product.ID.IsZero() {
		return nil, ErrParentProductNotFound
	}
	productVariation.CreatedBy = user.ID
	productVariation, err = models.CreateProductVariation(*productVariation)
	if err != nil {
		return nil, err
	}
	product.Variations = append(product.Variations, productVariation.ID)
	product, err = models.UpdateProduct(product)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), productVariation.ID.Hex(), "product variation", productVariation, nil, ctx)
	return productVariation, nil
}

// productVariationResolver is of type struct.
type productVariationResolver struct{ *Resolver }

func (r productVariationResolver) Weight(ctx context.Context, obj *models.ProductVariation) (float64, error) {
	return obj.Weight, nil
}

func (r *mutationResolver) UpdateProductVariation(ctx context.Context, input models.UpdateProductVariationInput) (*models.ProductVariation, error) {
	productVariation := &models.ProductVariation{}
	productVariation, err := models.GetProductVariationByID(input.ID)
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&productVariation, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	productVariation.CreatedBy = user.ID
	productVariation = models.UpdateProductVariation(productVariation)
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), productVariation.ID.Hex(), "product variation", productVariation, nil, ctx)
	return productVariation, nil
}

func (r *mutationResolver) DeleteProductVariation(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteProductVariationByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "product variation", nil, nil, ctx)
	return &res, err
}

func (r *queryResolver) ProductVariations(ctx context.Context, productVariationID *primitive.ObjectID, after *string, before *string, first *int, last *int) (*models.ProductVariationConnection, error) {
	var items []*models.ProductVariation
	var edges []*models.ProductVariationEdge
	filter := bson.D{}
	limit := 25

	items, totalCount, hasPrevious, hasNext, err := models.GetProductVariations(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.ProductVariationEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	productVariationList := &models.ProductVariationConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return productVariationList, nil
}

func (r *queryResolver) ProductVariation(ctx context.Context, id primitive.ObjectID) (*models.ProductVariation, error) {
	productVariation, err := models.GetProductVariationByID(id)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return productVariation, nil
}
