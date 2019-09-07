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

func (r *queryResolver) ProductDownloads(ctx context.Context, id primitive.ObjectID, after *string, before *string, first *int, last *int) (*models.ProductDownloadConnection, error) {
	var items []*models.ProductDownload
	var edges []*models.ProductDownloadEdge
	filter := bson.D{}
	limit := 25

	items, totalCount, hasPrevious, hasNext, err := models.GetProductDownloads(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.ProductDownloadEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	ProductDownloadList := &models.ProductDownloadConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return ProductDownloadList, nil
}

func (r *queryResolver) ProductDownload(ctx context.Context, id primitive.ObjectID) (*models.ProductDownload, error) {
	productDownload, err := models.GetProductDownloadByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return productDownload, nil
}

func (r *mutationResolver) AddProductDownload(ctx context.Context, input models.AddProductDownloadInput) (*models.ProductDownload, error) {
	productDownload := &models.ProductDownload{}
	_ = copier.Copy(&productDownload, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	productDownload.CreatedBy = user.ID
	productDownload, err = models.CreateProductDownload(*productDownload)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), productDownload.ID.Hex(), "product download", productDownload, nil, ctx)
	return productDownload, nil
}

func (r *mutationResolver) UpdateProductDownload(ctx context.Context, input models.UpdateProductDownloadInput) (*models.ProductDownload, error) {
	productDownload := &models.ProductDownload{}
	productDownload, err := models.GetProductDownloadByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&productDownload, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	productDownload.CreatedBy = user.ID
	productDownload = models.UpdateProductDownload(productDownload)
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), productDownload.ID.Hex(), "product download", productDownload, nil, ctx)
	return productDownload, nil
}

func (r *mutationResolver) DeleteProductDownload(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteProductDownloadByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "product download", nil, nil, ctx)
	return &res, err
}
