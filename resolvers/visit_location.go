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

//VisitLocations gives a list of visit locations
func (r *queryResolver) VisitLocations(ctx context.Context, visitLocationType *models.VisitLocationType, text *string, after *string, before *string, first *int, last *int) (*models.VisitLocationConnection, error) {
	var items []*models.VisitLocation
	var edges []*models.VisitLocationEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetVisitLocations(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.VisitLocationEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.VisitLocationConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//VisitLocation returns a visit location by ID
func (r *queryResolver) VisitLocation(ctx context.Context, id primitive.ObjectID) (*models.VisitLocation, error) {
	location, err := models.GetVisitLocationByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return location, nil
}

//AddVisitLocation adds a new visit location
func (r *mutationResolver) AddVisitLocation(ctx context.Context, input models.AddVisitLocationInput) (*models.VisitLocation, error) {
	location := &models.VisitLocation{}
	_ = copier.Copy(&location, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	location.CreatedBy = user.ID
	location, err = models.CreateVisitLocation(*location)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), location.ID.Hex(), "visit location", location, nil, ctx)
	return location, nil
}

//UpdateVisitLocation updates an existing visit location
func (r *mutationResolver) UpdateVisitLocation(ctx context.Context, input models.UpdateVisitLocationInput) (*models.VisitLocation, error) {
	location := &models.VisitLocation{}
	location, err := models.GetVisitLocationByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&location, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	location.CreatedBy = user.ID
	location, err = models.UpdateVisitLocation(location)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), location.ID.Hex(), "visit location", location, nil, ctx)
	return location, nil
}

//DeleteVisitLocation deletes visit location
func (r *mutationResolver) DeleteVisitLocation(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteVisitLocationByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "visit location", nil, nil, ctx)
	return &res, err
}

//ActivateVisitLocation activates a visit location
func (r *mutationResolver) ActivateVisitLocation(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	location, err := models.GetVisitLocationByID(id.Hex())
	if err != nil {
		return nil, err
	}
	location.IsActive = true
	_, err = models.UpdateVisitLocation(location)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "visit location", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//DeactivateVisitLocation deactivates a visit location
func (r *mutationResolver) DeactivateVisitLocation(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	location, err := models.GetVisitLocationByID(id.Hex())
	if err != nil {
		return nil, err
	}
	location.IsActive = false
	_, err = models.UpdateVisitLocation(location)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "visit location", nil, nil, ctx)
	return utils.PointerBool(false), nil
}

// visitLocationResolver is of type struct.
type visitLocationResolver struct{ *Resolver }
