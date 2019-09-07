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

//VehicleMake returns a vehicle make by ID
func (r *queryResolver) VehicleMake(ctx context.Context, id primitive.ObjectID) (*models.VehicleMake, error) {
	vehicleMake, err := models.GetVehicleMakeByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return vehicleMake, nil
}

type vehicleMakeResolver struct{ *Resolver }

//VehicleMakes gives a list of vehicle makes
func (r *queryResolver) VehicleMakes(ctx context.Context, vehicleMakeType *models.VehicleMakeType, text *string, after *string, before *string, first *int, last *int) (*models.VehicleMakeConnection, error) {
	var items []*models.VehicleMake
	var edges []*models.VehicleMakeEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetVehicleMakes(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.VehicleMakeEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.VehicleMakeConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//AddVehicleMake adds a new vehicle make
func (r *mutationResolver) AddVehicleMake(ctx context.Context, input models.AddVehicleMakeInput) (*models.VehicleMake, error) {
	vehicleMake := &models.VehicleMake{}
	_ = copier.Copy(&vehicleMake, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	vehicleMake.CreatedBy = user.ID
	vehicleMake, err = models.CreateVehicleMake(*vehicleMake)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), vehicleMake.ID.Hex(), "vehicle make", vehicleMake, nil, ctx)
	return vehicleMake, nil
}

//UpdateVehicleMake updates an existing vehicle make
func (r *mutationResolver) UpdateVehicleMake(ctx context.Context, input models.UpdateVehicleMakeInput) (*models.VehicleMake, error) {
	vehicleMake := &models.VehicleMake{}
	vehicleMake, err := models.GetVehicleMakeByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&vehicleMake, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	vehicleMake.CreatedBy = user.ID
	vehicleMake, err = models.UpdateVehicleMake(vehicleMake)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), vehicleMake.ID.Hex(), "vehicle make", vehicleMake, nil, ctx)
	return vehicleMake, nil
}

//DeleteVehicleMake deletes an existing vehicle make
func (r *mutationResolver) DeleteVehicleMake(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteVehicleMakeByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "vehicle make", nil, nil, ctx)
	return &res, err
}

//ActivateVehicleMake activates vehicle make by its ID
func (r *mutationResolver) ActivateVehicleMake(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	vehicleMake, err := models.GetVehicleMakeByID(id.Hex())
	if err != nil {
		return nil, err
	}
	vehicleMake.IsActive = true
	_, err = models.UpdateVehicleMake(vehicleMake)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "vehicle make", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//DeactivateVehicleMake deactivates vehicle make by its ID
func (r *mutationResolver) DeactivateVehicleMake(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	vehicleMake, err := models.GetVehicleMakeByID(id.Hex())
	if err != nil {
		return nil, err
	}
	vehicleMake.IsActive = false
	_, err = models.UpdateVehicleMake(vehicleMake)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "vehicle make", nil, nil, ctx)
	return utils.PointerBool(false), nil
}
