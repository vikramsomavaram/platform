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

type rentalPackageResolver struct{ *Resolver }

func (r *queryResolver) RentalPackage(ctx context.Context, id primitive.ObjectID) (*models.RentalPackage, error) {
	rentalPackage, err := models.GetRentalPackageByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return rentalPackage, nil
}

func (r *queryResolver) RentalPackages(ctx context.Context, appID primitive.ObjectID, after *string, before *string, first *int, last *int) (*models.RentalPackageConnection, error) {
	var items []*models.RentalPackage
	var edges []*models.RentalPackageEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetRentalPackages(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.RentalPackageEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.RentalPackageConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

func (r *mutationResolver) AddRentalPackage(ctx context.Context, input models.AddRentalPackageInput) (*models.RentalPackage, error) {
	rentalPackage := &models.RentalPackage{}
	_ = copier.Copy(&rentalPackage, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	rentalPackage.CreatedBy = user.ID
	rentalPackage, err = models.CreateRentalPackage(*rentalPackage)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), rentalPackage.ID.Hex(), "rental package", rentalPackage, nil, ctx)
	return rentalPackage, nil
}

func (r *mutationResolver) UpdateRentalPackage(ctx context.Context, input models.UpdateRentalPackageInput) (*models.RentalPackage, error) {
	rentalPackage := &models.RentalPackage{}
	rentalPackage, err := models.GetRentalPackageByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&rentalPackage, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	rentalPackage.CreatedBy = user.ID
	rentalPackage, err = models.UpdateRentalPackage(rentalPackage)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), rentalPackage.ID.Hex(), "rental package", rentalPackage, nil, ctx)
	return rentalPackage, nil
}

func (r *mutationResolver) DeleteRentalPackage(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteRentalPackageByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "rental package", nil, nil, ctx)
	return &res, err
}
