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

//AddAirportSurcharge adds a new airport surcharge
func (r *mutationResolver) AddAirportSurcharge(ctx context.Context, input models.AddAirportSurchargeInput) (*models.AirportSurcharge, error) {
	surcharge := &models.AirportSurcharge{}
	_ = copier.Copy(&surcharge, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	surcharge.CreatedBy = user.ID
	surcharge, err = models.CreateAirportSurcharge(*surcharge)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), surcharge.ID.Hex(), "surcharge", surcharge, nil, ctx)
	return surcharge, nil
}

//UpdateAirportSurcharge updates an existin surcharge
func (r *mutationResolver) UpdateAirportSurcharge(ctx context.Context, input models.UpdateAirportSurchargeInput) (*models.AirportSurcharge, error) {
	surcharge := &models.AirportSurcharge{}
	surcharge, err := models.GetAirportSurchargeByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&surcharge, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	surcharge.CreatedBy = user.ID
	surcharge, err = models.UpdateAirportSurcharge(surcharge)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), surcharge.ID.Hex(), "surcharge", surcharge, nil, ctx)
	return surcharge, nil
}

//DeleteAirportSurcharge deletes an airport surcharge
func (r *mutationResolver) DeleteAirportSurcharge(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteAirportSurchargeByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "surcharge", nil, nil, ctx)
	return &res, err
}

//ActivateAirportSurcharge activates an airport surcharge by ID
func (r *mutationResolver) ActivateAirportSurcharge(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	surcharge, err := models.GetAirportSurchargeByID(id.Hex())
	if err != nil {
		return nil, err
	}
	surcharge.IsActive = true
	_, err = models.UpdateAirportSurcharge(surcharge)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "surcharge", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//DeactivateAirportSurcharge deactivates an airport surcharge by ID
func (r *mutationResolver) DeactivateAirportSurcharge(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	surcharge, err := models.GetAirportSurchargeByID(id.Hex())
	if err != nil {
		return nil, err
	}
	surcharge.IsActive = false
	_, err = models.UpdateAirportSurcharge(surcharge)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "surcharge", nil, nil, ctx)
	return utils.PointerBool(false), nil
}

//AirportSurcharges gives a list of airport surcharges
func (r *queryResolver) AirportSurcharges(ctx context.Context, airportAirportSurchargeSearch *models.AirportSurchargeSearch, text *string, after *string, before *string, first *int, last *int) (*models.AirportSurchargeConnection, error) {
	var items []*models.AirportSurcharge
	var edges []*models.AirportSurchargeEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetAirportSurcharges(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.AirportSurchargeEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.AirportSurchargeConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//AirportSurcharge returns an airport surcharge by its ID
func (r *queryResolver) AirportSurcharge(ctx context.Context, id primitive.ObjectID) (*models.AirportSurcharge, error) {
	surcharge, err := models.GetAirportSurchargeByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return surcharge, nil
}

type airportSurchargeResolver struct{ *Resolver }
