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
	"time"
)

//GeoFenceLocations gives a list of geo fence locations
func (r *queryResolver) GeoFenceLocations(ctx context.Context, geoFenceLocationType *models.GeoFenceLocationSearchType, text *string, geoFenceLocationStatus *models.GeoFenceLocationStatus, after *string, before *string, first *int, last *int) (*models.GeoFenceLocationConnection, error) {
	var items []*models.GeoFenceLocation
	var edges []*models.GeoFenceLocationEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetGeoFenceLocations(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.GeoFenceLocationEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.GeoFenceLocationConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//GeoFenceLocation returns a geo fence location by its ID
func (r *queryResolver) GeoFenceLocation(ctx context.Context, id primitive.ObjectID) (*models.GeoFenceLocation, error) {
	location, err := models.GetGeoFenceLocationByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return location, nil
}

type geoFenceLocationResolver struct{ *Resolver }

//AddGeoFenceLocation adds a new geo fence location
func (r *mutationResolver) AddGeoFenceLocation(ctx context.Context, input models.AddGeoFenceLocationInput) (*models.GeoFenceLocation, error) {
	location := &models.GeoFenceLocation{}
	_ = copier.Copy(&location, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	location.CreatedBy = user.ID
	location, err = models.CreateGeoFenceLocation(*location)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), location.ID.Hex(), "geo fence location", location, nil, ctx)
	return location, nil
}

//UpdateGeoFenceLocation updates an existing geo fence location
func (r *mutationResolver) UpdateGeoFenceLocation(ctx context.Context, input models.UpdateGeoFenceLocationInput) (*models.GeoFenceLocation, error) {
	location := &models.GeoFenceLocation{}
	location, err := models.GetGeoFenceLocationByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&location, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	location.CreatedBy = user.ID
	location.UpdatedAt = time.Now()
	location, err = models.UpdateGeoFenceLocation(location)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), location.ID.Hex(), "geo fence location", location, nil, ctx)
	return location, nil
}

//DeleteGeoFenceLocation deletes an existing geo fence location
func (r *mutationResolver) DeleteGeoFenceLocation(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteGeoFenceLocationByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "geo fence location", nil, nil, ctx)
	return &res, err
}

//ActivateGeoFenceLocation activates a geo fence location by its ID
func (r *mutationResolver) ActivateGeoFenceLocation(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	location, err := models.GetGeoFenceLocationByID(id.Hex())
	if err != nil {
		return nil, err
	}
	location.IsActive = true
	_, err = models.UpdateGeoFenceLocation(location)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "geo fence location", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//DeactivateGeoFenceLocation deactivates a geo fence location by its ID
func (r *mutationResolver) DeactivateGeoFenceLocation(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	location, err := models.GetGeoFenceLocationByID(id.Hex())
	if err != nil {
		return nil, err
	}
	location.IsActive = false
	_, err = models.UpdateGeoFenceLocation(location)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "geo fence location", nil, nil, ctx)
	return utils.PointerBool(false), nil
}
