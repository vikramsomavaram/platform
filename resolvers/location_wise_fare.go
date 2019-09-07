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

//AddLocationWiseFare adds a new location wise fare
func (r *mutationResolver) AddLocationWiseFare(ctx context.Context, input models.AddLocationWiseFareInput) (*models.LocationWiseFare, error) {
	locationWiseFare := &models.LocationWiseFare{}
	_ = copier.Copy(&locationWiseFare, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	locationWiseFare.CreatedBy = user.ID
	locationWiseFare, err = models.CreateLocationWiseFare(*locationWiseFare)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), locationWiseFare.ID.Hex(), "location wise fare", locationWiseFare, nil, ctx)
	return locationWiseFare, nil
}

//UpdateLocationWiseFare updates an existing location wise fare
func (r *mutationResolver) UpdateLocationWiseFare(ctx context.Context, input models.UpdateLocationWiseFareInput) (*models.LocationWiseFare, error) {
	locationWiseFare := &models.LocationWiseFare{}
	locationWiseFare, err := models.GetLocationWiseFareByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&locationWiseFare, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	locationWiseFare.CreatedBy = user.ID
	locationWiseFare, err = models.UpdateLocationWiseFare(locationWiseFare)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), locationWiseFare.ID.Hex(), "location wise fare", locationWiseFare, nil, ctx)
	return locationWiseFare, nil
}

//DeleteLocationWiseFare deletes a location wise fare
func (r *mutationResolver) DeleteLocationWiseFare(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteLocationWiseFareByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "location wise fare", nil, nil, ctx)
	return &res, err
}

//ActivateLocationWiseFare activates a location wise fare by its ID
func (r *mutationResolver) ActivateLocationWiseFare(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	locationWiseFare, err := models.GetLocationWiseFareByID(id.Hex())
	if err != nil {
		return nil, err
	}
	locationWiseFare.IsActive = true
	_, err = models.UpdateLocationWiseFare(locationWiseFare)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "location wise fare", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//DeactivateLocationWiseFare deactivates a location wise fare by its ID
func (r *mutationResolver) DeactivateLocationWiseFare(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	locationWiseFare, err := models.GetLocationWiseFareByID(id.Hex())
	if err != nil {
		return nil, err
	}
	locationWiseFare.IsActive = false
	_, err = models.UpdateLocationWiseFare(locationWiseFare)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "location wise fare", nil, nil, ctx)
	return utils.PointerBool(false), nil
}

//LocationWiseFares gives a list of location wise fares
func (r *queryResolver) LocationWiseFares(ctx context.Context, locationFareSearch *models.LocationWiseFareSearch, text *string, after *string, before *string, first *int, last *int) (*models.LocationWiseFareConnection, error) {
	var items []*models.LocationWiseFare
	var edges []*models.LocationWiseFareEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetLocationWiseFares(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.LocationWiseFareEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.LocationWiseFareConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//LocationWiseFare returns a location wise fare by its ID
func (r *queryResolver) LocationWiseFare(ctx context.Context, id primitive.ObjectID) (*models.LocationWiseFare, error) {
	locationWiseFare, err := models.GetLocationWiseFareByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return locationWiseFare, nil
}

type locationWiseFareResolver struct{ *Resolver }
