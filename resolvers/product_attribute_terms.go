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

func (r *queryResolver) ProductAttributeTerms(ctx context.Context, id primitive.ObjectID, after *string, before *string, first *int, last *int) (*models.ProductAttributeTermConnection, error) {
	var items []*models.ProductAttributeTerm
	var edges []*models.ProductAttributeTermEdge
	filter := bson.D{}
	limit := 25

	items, totalCount, hasPrevious, hasNext, err := models.GetProductAttributeTerms(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.ProductAttributeTermEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	ProductAttributeTermList := &models.ProductAttributeTermConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return ProductAttributeTermList, nil
}

func (r *queryResolver) ProductAttributeTerm(ctx context.Context, id primitive.ObjectID) (*models.ProductAttributeTerm, error) {
	productAttributeTerm := models.GetProductAttributeTermByID(id.Hex())
	return productAttributeTerm, nil
}

func (r *mutationResolver) AddProductAttributeTerm(ctx context.Context, input models.AddProductAttributeTermInput) (*models.ProductAttributeTerm, error) {
	productAttributeTerm := &models.ProductAttributeTerm{}
	_ = copier.Copy(&productAttributeTerm, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	productAttributeTerm.CreatedBy = user.ID
	productAttributeTerm, err = models.CreateProductAttributeTerm(*productAttributeTerm)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), productAttributeTerm.ID.Hex(), "product attribute term", productAttributeTerm, nil, ctx)
	return productAttributeTerm, nil
}

func (r *mutationResolver) UpdateProductAttributeTerm(ctx context.Context, input models.UpdateProductAttributeTermInput) (*models.ProductAttributeTerm, error) {
	productAttributeTerm := &models.ProductAttributeTerm{}
	productAttributeTerm = models.GetProductAttributeTermByID(input.ID.Hex())
	_ = copier.Copy(&productAttributeTerm, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	productAttributeTerm.CreatedBy = user.ID
	productAttributeTerm = models.UpdateProductAttributeTerm(productAttributeTerm)
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), productAttributeTerm.ID.Hex(), "product attribute term", productAttributeTerm, nil, ctx)
	return productAttributeTerm, nil
}

func (r *mutationResolver) DeleteProductAttributeTerm(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteProductAttributeTermByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "product attribute term", nil, nil, ctx)
	return &res, err
}
