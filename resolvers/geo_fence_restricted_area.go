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

//AddGeoFenceRestrictedArea adds a new geo fence restricted area
func (r *mutationResolver) AddGeoFenceRestrictedArea(ctx context.Context, input models.AddGeoFenceRestrictedAreaInput) (*models.GeoFenceRestrictedArea, error) {
	restrictedArea := &models.GeoFenceRestrictedArea{}
	_ = copier.Copy(&restrictedArea, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	restrictedArea.CreatedBy = user.ID
	restrictedArea, err = models.CreateGeoFenceRestrictedArea(*restrictedArea)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), restrictedArea.ID.Hex(), "geo fence restricted area", restrictedArea, nil, ctx)
	return restrictedArea, nil
}

//UpdateGeoFenceRestrictedArea updates an existing geo fence restricted area
func (r *mutationResolver) UpdateGeoFenceRestrictedArea(ctx context.Context, input models.UpdateGeoFenceRestrictedAreaInput) (*models.GeoFenceRestrictedArea, error) {
	restrictedArea := &models.GeoFenceRestrictedArea{}
	restrictedArea, err := models.GetGeoFenceRestrictedAreaByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&restrictedArea, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	restrictedArea.CreatedBy = user.ID
	restrictedArea, err = models.UpdateGeoFenceRestrictedArea(restrictedArea)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), restrictedArea.ID.Hex(), "geo fence restricted area", restrictedArea, nil, ctx)
	return restrictedArea, nil
}

//DeleteGeoFenceRestrictedArea deletes an existing geo fence restricted area
func (r *mutationResolver) DeleteGeoFenceRestrictedArea(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteGeoFenceRestrictedAreaByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "geo fence restricted area", nil, nil, ctx)
	return &res, err
}

//ActivateGeoFenceRestrictedArea activates a geo fence restricted area by its ID
func (r *mutationResolver) ActivateGeoFenceRestrictedArea(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	restrictedArea, err := models.GetGeoFenceRestrictedAreaByID(id.Hex())
	if err != nil {
		return nil, err
	}
	restrictedArea.IsActive = true
	_, err = models.UpdateGeoFenceRestrictedArea(restrictedArea)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "geo fence restricted area", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//DeactivateGeoFenceRestrictedArea deactivates a geo fence restricted area by its ID
func (r *mutationResolver) DeactivateGeoFenceRestrictedArea(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	restrictedArea, err := models.GetGeoFenceRestrictedAreaByID(id.Hex())
	if err != nil {
		return nil, err
	}
	restrictedArea.IsActive = false
	_, err = models.UpdateGeoFenceRestrictedArea(restrictedArea)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "geo fence restricted area", nil, nil, ctx)
	return utils.PointerBool(false), nil
}

//GeoFenceRestrictedAreas gives a list of geo fence restricted areas
func (r *queryResolver) GeoFenceRestrictedAreas(ctx context.Context, geoFenceRestrictedAreaType *models.GeoFenceRestrictedAreaSearchType, text *string, geoFenceRestrictedAreaStatus *models.GeoFenceRestrictedAreaStatus, after *string, before *string, first *int, last *int) (*models.GeoFenceRestrictedAreaConnection, error) {
	var items []*models.GeoFenceRestrictedArea
	var edges []*models.GeoFenceRestrictedAreaEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetGeoFenceRestrictedAreas(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.GeoFenceRestrictedAreaEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.GeoFenceRestrictedAreaConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//GeoFenceRestrictedArea returns a geo fence restricted area by its ID
func (r *queryResolver) GeoFenceRestrictedArea(ctx context.Context, id primitive.ObjectID) (*models.GeoFenceRestrictedArea, error) {
	restrictedArea, err := models.GetGeoFenceRestrictedAreaByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return restrictedArea, nil
}
