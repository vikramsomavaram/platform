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

//VehicleModels gives a list of vehicle models
func (r *queryResolver) VehicleModels(ctx context.Context, vehicleModelType *models.VehicleModelSearchType, text *string, after *string, before *string, first *int, last *int) (*models.VehicleModelConnection, error) {
	var items []*models.VehicleModel
	var edges []*models.VehicleModelEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetVehicleModels(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.VehicleModelEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.VehicleModelConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

type vehicleModelResolver struct{ *Resolver }

//VehicleModel returns a vehicle model by its ID
func (r *queryResolver) VehicleModel(ctx context.Context, id primitive.ObjectID) (*models.VehicleModel, error) {
	vehicleModel, err := models.GetVehicleModelByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return vehicleModel, nil
}

//AddVehicleModel adds a new vehicle model
func (r *mutationResolver) AddVehicleModel(ctx context.Context, input models.AddVehicleModelInput) (*models.VehicleModel, error) {
	vehicleModel := &models.VehicleModel{}
	_ = copier.Copy(&vehicleModel, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	vehicleModel.CreatedBy = user.ID
	vehicleModel, err = models.CreateVehicleModel(*vehicleModel)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), vehicleModel.ID.Hex(), "vehicle model", vehicleModel, nil, ctx)
	return vehicleModel, nil
}

//UpdateVehicleModel updates an existing vehicle model
func (r *mutationResolver) UpdateVehicleModel(ctx context.Context, input models.UpdateVehicleModelInput) (*models.VehicleModel, error) {
	vehicleModel := &models.VehicleModel{}
	vehicleModel, err := models.GetVehicleModelByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&vehicleModel, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	vehicleModel.CreatedBy = user.ID
	vehicleModel, err = models.UpdateVehicleModel(vehicleModel)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), vehicleModel.ID.Hex(), "vehicle model", vehicleModel, nil, ctx)
	return vehicleModel, nil
}

//DeleteVehicleModel deletes an existing vehicle model
func (r *mutationResolver) DeleteVehicleModel(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteVehicleModelByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "vehicle model", nil, nil, ctx)
	return &res, err
}

//ActivateVehicleModel activates a vehicle model by ID
func (r *mutationResolver) ActivateVehicleModel(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	vehicleModel, err := models.GetVehicleModelByID(id.Hex())
	if err != nil {
		return nil, err
	}
	vehicleModel.IsActive = true
	_, err = models.UpdateVehicleModel(vehicleModel)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "vehicle model", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//DeactivateVehicleModel deactivates a vehicle model by ID
func (r *mutationResolver) DeactivateVehicleModel(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	vehicleModel, err := models.GetVehicleModelByID(id.Hex())
	if err != nil {
		return nil, err
	}
	vehicleModel.IsActive = false
	_, err = models.UpdateVehicleModel(vehicleModel)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "vehicle model", nil, nil, ctx)
	return utils.PointerBool(false), nil
}
