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

type storeVehicleTypeResolver struct{ *Resolver }

//StoreVehicleType returns a store vehicle type by its ID
func (r *queryResolver) StoreVehicleType(ctx context.Context, id primitive.ObjectID) (*models.StoreVehicleType, error) {
	storeVehicleType, err := models.GetStoreVehicleTypeByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return storeVehicleType, nil
}

//StoreVehicleTypes returns a list of store vehicle types
func (r *queryResolver) StoreVehicleTypes(ctx context.Context, storeVehicleTypeSearch *models.StoreVehicleTypeSearch, storeVehicleTypeStatus *models.StoreVehicleTypeStatus, storeVehicleTypeLocation *models.StoreVehicleTypeLocation, text *string, appID primitive.ObjectID, after *string, before *string, first *int, last *int) (*models.StoreVehicleTypeConnection, error) {
	var items []*models.StoreVehicleType
	var edges []*models.StoreVehicleTypeEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetStoreVehicleTypes(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.StoreVehicleTypeEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.StoreVehicleTypeConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//AddStoreVehicleType adds a new store vehicle type
func (r *mutationResolver) AddStoreVehicleType(ctx context.Context, input models.AddStoreVehicleTypeInput) (*models.StoreVehicleType, error) {
	storeVehicleType := &models.StoreVehicleType{}
	_ = copier.Copy(&storeVehicleType, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	storeVehicleType.CreatedBy = user.ID
	storeVehicleType, err = models.CreateStoreVehicleType(*storeVehicleType)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), storeVehicleType.ID.Hex(), "store vehicle type", storeVehicleType, nil, ctx)
	return storeVehicleType, nil
}

//UpdateStoreVehicleType updates an existing store vehicle type
func (r *mutationResolver) UpdateStoreVehicleType(ctx context.Context, input models.UpdateStoreVehicleTypeInput) (*models.StoreVehicleType, error) {
	storeVehicleType := &models.StoreVehicleType{}
	storeVehicleType, err := models.GetStoreVehicleTypeByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&storeVehicleType, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	storeVehicleType.CreatedBy = user.ID
	storeVehicleType, err = models.UpdateStoreVehicleType(storeVehicleType)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), storeVehicleType.ID.Hex(), "store vehicle type", storeVehicleType, nil, ctx)
	return storeVehicleType, nil
}

//DeleteStoreVehicleType deletes a store vehicle type
func (r *mutationResolver) DeleteStoreVehicleType(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteStoreVehicleTypeByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "store vehicle type", nil, nil, ctx)
	return &res, err
}

//ActivateStoreVehicleType activates a store vehicle type by its ID
func (r *mutationResolver) ActivateStoreVehicleType(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	storeVehicleType, err := models.GetStoreVehicleTypeByID(id.Hex())
	if err != nil {
		return nil, err
	}
	storeVehicleType.IsActive = true
	_, err = models.UpdateStoreVehicleType(storeVehicleType)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "store vehicle type", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//DeactivateStoreVehicleType deactivates a store vehicle type by its ID
func (r *mutationResolver) DeactivateStoreVehicleType(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	storeVehicleType, err := models.GetStoreVehicleTypeByID(id.Hex())
	if err != nil {
		return nil, err
	}
	storeVehicleType.IsActive = false
	_, err = models.UpdateStoreVehicleType(storeVehicleType)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "store vehicle type", nil, nil, ctx)
	return utils.PointerBool(false), nil
}
