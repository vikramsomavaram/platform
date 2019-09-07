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
	"github.com/tribehq/platform/utils"
	"github.com/tribehq/platform/utils/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//PackageType returns a given package type by its ID
func (r *queryResolver) PackageType(ctx context.Context, packageTypeID string) (*models.PackageType, error) {
	packageType, err := models.GetPackageTypeByID(packageTypeID)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return packageType, nil
}

//PackageTypes returns a list package types.
func (r *queryResolver) PackageTypes(ctx context.Context, searchPackageType *models.SearchPackageType, text *string, searchPackageTypeStatus *models.SearchPackageTypeStatus, after *string, before *string, first *int, last *int) (*models.PackageTypeConnection, error) {
	var items []*models.PackageType
	var edges []*models.PackageTypeEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetPackageTypes(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.PackageTypeEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.PackageTypeConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//UpdatePackageType updates an existing package type
func (r *mutationResolver) UpdatePackageType(ctx context.Context, input models.UpdatePackageTypeInput) (*models.PackageType, error) {
	packageType := &models.PackageType{}
	packageType, err := models.GetPackageTypeByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&packageType, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	packageType.CreatedBy = user.ID
	packageType, err = models.UpdatePackageType(packageType)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), packageType.ID.Hex(), "package type", packageType, nil, ctx)
	return packageType, nil
}

//AddPackageType adds a new package type
func (r *mutationResolver) AddPackageType(ctx context.Context, input models.AddPackageTypeInput) (*models.PackageType, error) {
	packageType := &models.PackageType{}
	_ = copier.Copy(&packageType, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	packageType.CreatedBy = user.ID
	packageType, err = models.CreatePackageType(*packageType)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), packageType.ID.Hex(), "package type", packageType, nil, ctx)
	return packageType, nil
}

//DeletePackageType deletes an existing package type
func (r *mutationResolver) DeletePackageType(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeletePackageTypeByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "package type", nil, nil, ctx)
	return &res, err
}

//ActivatePackageType activates a package type by its ID
func (r *mutationResolver) ActivatePackageType(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	packageType, err := models.GetPackageTypeByID(id.Hex())
	if err != nil {
		return nil, err
	}
	packageType.IsActive = true
	_, err = models.UpdatePackageType(packageType)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "package type", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//DeactivatePackageType deactivates a package type by its ID
func (r *mutationResolver) DeactivatePackageType(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	packageType, err := models.GetPackageTypeByID(id.Hex())
	if err != nil {
		return nil, err
	}
	packageType.IsActive = false
	_, err = models.UpdatePackageType(packageType)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "package type", nil, nil, ctx)
	return utils.PointerBool(false), nil
}

// packageTypeResolver is of type struct of package type.
type packageTypeResolver struct{ *Resolver }

func (packageTypeResolver) ID(ctx context.Context, obj *models.PackageType) (*string, error) {
	id := obj.ID.Hex()
	return &id, nil
}
