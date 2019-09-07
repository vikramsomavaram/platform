/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package resolvers

import (
	"encoding/base64"
	"github.com/jinzhu/copier"
	log "github.com/sirupsen/logrus"
	"github.com/tribehq/platform/lib/audit_log"
	"github.com/tribehq/platform/models"
	"github.com/tribehq/platform/utils"
	"github.com/tribehq/platform/utils/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/net/context"
)

//RideProfileTypes gives a list of ride profile types
func (r *queryResolver) RideProfileTypes(ctx context.Context, rideProfileType *models.RideProfileSearchType, text *string, after *string, before *string, first *int, last *int) (*models.RideProfileTypeConnection, error) {
	var items []*models.RideProfileType
	var edges []*models.RideProfileTypeEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetRideProfileTypes(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.RideProfileTypeEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.RideProfileTypeConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//RideProfileType returns a ride profile type by its ID
func (r *queryResolver) RideProfileType(ctx context.Context, id primitive.ObjectID) (*models.RideProfileType, error) {
	rideProfileType, err := models.GetRideProfileTypeByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return rideProfileType, nil
}

//AddRideProfileType adds a new ride profile type
func (r *mutationResolver) AddRideProfileType(ctx context.Context, input models.AddRideProfileTypeInput) (*models.RideProfileType, error) {
	rideProfileType := &models.RideProfileType{}
	_ = copier.Copy(&rideProfileType, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	rideProfileType.CreatedBy = user.ID
	rideProfileType, err = models.CreateRideProfileType(*rideProfileType)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), rideProfileType.ID.Hex(), "ride profile type", rideProfileType, nil, ctx)
	return rideProfileType, nil
}

//UpdateRideProfileType updates an existing ride profile type
func (r *mutationResolver) UpdateRideProfileType(ctx context.Context, input models.UpdateRideProfileTypeInput) (*models.RideProfileType, error) {
	rideProfileType := &models.RideProfileType{}
	rideProfileType, err := models.GetRideProfileTypeByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&rideProfileType, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	rideProfileType.CreatedBy = user.ID
	rideProfileType, err = models.UpdateRideProfileType(rideProfileType)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), rideProfileType.ID.Hex(), "ride profile type", rideProfileType, nil, ctx)
	return rideProfileType, nil
}

//DeleteRideProfileType deletes an existing ride profile type
func (r *mutationResolver) DeleteRideProfileType(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteRideProfileTypeByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "ride profile type", nil, nil, ctx)
	return &res, err
}

//ActivateRideProfileType activates a ride profile type by its ID
func (r *mutationResolver) ActivateRideProfileType(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	rideProfileType, err := models.GetRideProfileTypeByID(id.Hex())
	if err != nil {
		return nil, err
	}
	rideProfileType.IsActive = true
	_, err = models.UpdateRideProfileType(rideProfileType)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "ride profile type", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//DeactivateRideProfileType deactivates a ride profile type by its ID
func (r *mutationResolver) DeactivateRideProfileType(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	rideProfileType, err := models.GetRideProfileTypeByID(id.Hex())
	if err != nil {
		return nil, err
	}
	rideProfileType.IsActive = false
	_, err = models.UpdateRideProfileType(rideProfileType)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "ride profile type", nil, nil, ctx)
	return utils.PointerBool(false), nil
}

//rideProfileTypeResolver is of type struct.
type rideProfileTypeResolver struct{ *Resolver }
