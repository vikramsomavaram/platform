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

func (r *mutationResolver) AddProductReview(ctx context.Context, input models.AddProductReviewInput) (*models.ProductReview, error) {
	productReview := &models.ProductReview{}
	_ = copier.Copy(&productReview, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	productReview.CreatedBy = user.ID
	productReview, err = models.CreateProductReview(*productReview)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), productReview.ID.Hex(), "product review", productReview, nil, ctx)
	return productReview, nil
}

func (r *mutationResolver) UpdateProductReview(ctx context.Context, input models.UpdateProductReviewInput) (*models.ProductReview, error) {
	productReview := &models.ProductReview{}
	productReview = models.GetProductReviewByID(input.ID.Hex())
	_ = copier.Copy(&productReview, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	productReview.CreatedBy = user.ID
	productReview = models.UpdateProductReview(productReview)
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), productReview.ID.Hex(), "product review", productReview, nil, ctx)
	return productReview, nil
}

// productReviewResolver is of type struct.
type productReviewResolver struct{ *Resolver }

func (r productReviewResolver) ProductID(ctx context.Context, obj *models.ProductReview) (string, error) {
	return obj.ProductID.Hex(), nil
}

func (r *mutationResolver) DeleteProductReview(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteProductReviewByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "product review", nil, nil, ctx)
	return &res, err
}

func (r *queryResolver) ProductReviews(ctx context.Context, productReviewID *primitive.ObjectID, after *string, before *string, first *int, last *int) (*models.ProductReviewConnection, error) {
	var items []*models.ProductReview
	var edges []*models.ProductReviewEdge
	filter := bson.D{}
	limit := 25

	items, totalCount, hasPrevious, hasNext, err := models.GetProductReviews(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.ProductReviewEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	ProductReviewList := &models.ProductReviewConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return ProductReviewList, nil
}

func (r *queryResolver) ProductReview(ctx context.Context, id primitive.ObjectID) (*models.ProductReview, error) {
	productReview := models.GetProductReviewByID(id.Hex())
	return productReview, nil
}
