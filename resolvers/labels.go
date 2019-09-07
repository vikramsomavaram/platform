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
	"github.com/tribehq/platform/utils/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//GeneralLabels gives a list of general labels
func (r *queryResolver) GeneralLabels(ctx context.Context, generalLabelSearch *models.GeneralLabelSearch, text *string, after *string, before *string, first *int, last *int) (*models.GeneralLabelConnection, error) {
	var items []*models.GeneralLabel
	var edges []*models.GeneralLabelEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetGeneralLabels(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.GeneralLabelEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.GeneralLabelConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//GeneralLabel returns a general label by its ID
func (r *queryResolver) GeneralLabel(ctx context.Context, id primitive.ObjectID) (*models.GeneralLabel, error) {
	label, err := models.GetGeneralLabelByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return label, nil
}

//AddGeneralLabel adds a new general label
func (r *mutationResolver) AddGeneralLabel(ctx context.Context, input models.AddGeneralLabelInput) (*models.GeneralLabel, error) {
	label := &models.GeneralLabel{}
	_ = copier.Copy(&label, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	label.CreatedBy = user.ID
	label, err = models.CreateGeneralLabel(*label)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), label.ID.Hex(), "general label", label, nil, ctx)
	return label, nil
}

//UpdateGeneralLabel updates an existing general label
func (r *mutationResolver) UpdateGeneralLabel(ctx context.Context, input models.UpdateGeneralLabelInput) (*models.GeneralLabel, error) {
	label := &models.GeneralLabel{}
	label, err := models.GetGeneralLabelByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&label, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	label.CreatedBy = user.ID
	label, err = models.UpdateGeneralLabel(label)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), label.ID.Hex(), "general label", label, nil, ctx)
	return label, nil
}

//DeleteGeneralLabel deletes an existing general label
func (r *mutationResolver) DeleteGeneralLabel(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteGeneralLabelByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "general label", nil, nil, ctx)
	return &res, err
}

//AddFoodDeliveryLabel adds a new food delivery label
func (r *mutationResolver) AddFoodDeliveryLabel(ctx context.Context, input models.AddFoodDeliveryLabelInput) (*models.FoodDeliveryLabel, error) {
	label := &models.FoodDeliveryLabel{}
	_ = copier.Copy(&label, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	label.CreatedBy = user.ID
	label, err = models.CreateFoodDeliveryLabel(*label)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), label.ID.Hex(), "food delivery label", label, nil, ctx)
	return label, nil
}

//UpdateFoodDeliveryLabel updates an existing food delivery label
func (r *mutationResolver) UpdateFoodDeliveryLabel(ctx context.Context, input models.UpdateFoodDeliveryLabelInput) (*models.FoodDeliveryLabel, error) {
	label := &models.FoodDeliveryLabel{}
	label, err := models.GetFoodDeliveryLabelByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&label, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	label.CreatedBy = user.ID
	label, err = models.UpdateFoodDeliveryLabel(label)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), label.ID.Hex(), "food delivery label", label, nil, ctx)
	return label, nil
}

//DeleteFoodDeliveryLabel deletes an existing food delivery label
func (r *mutationResolver) DeleteFoodDeliveryLabel(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteFoodDeliveryLabelByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "food delivery label", nil, nil, ctx)
	return &res, err
}

//FoodDeliveryLabels gives a list of food delivery labels
func (r *queryResolver) FoodDeliveryLabels(ctx context.Context, foodDeliveryLabelSearch *models.FoodDeliveryLabelSearch, text *string, after *string, before *string, first *int, last *int) (*models.FoodDeliveryLabelConnection, error) {
	var items []*models.FoodDeliveryLabel
	var edges []*models.FoodDeliveryLabelEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetFoodDeliveryLabels(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.FoodDeliveryLabelEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.FoodDeliveryLabelConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//FoodDeliveryLabel returns a food delivery label by ID
func (r *queryResolver) FoodDeliveryLabel(ctx context.Context, id primitive.ObjectID) (*models.FoodDeliveryLabel, error) {
	label, err := models.GetFoodDeliveryLabelByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return label, nil
}

//GroceryDeliveryLabels gives a list of grocery delivery labels
func (r *queryResolver) GroceryDeliveryLabels(ctx context.Context, groceryDeliveryLabelSearch *models.GroceryDeliveryLabelSearch, text *string, after *string, before *string, first *int, last *int) (*models.GroceryDeliveryLabelConnection, error) {
	var items []*models.GroceryDeliveryLabel
	var edges []*models.GroceryDeliveryLabelEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetGroceryDeliveryLabels(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.GroceryDeliveryLabelEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.GroceryDeliveryLabelConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//GroceryDeliveryLabel returns a grocery delivery label by its ID
func (r *queryResolver) GroceryDeliveryLabel(ctx context.Context, id primitive.ObjectID) (*models.GroceryDeliveryLabel, error) {
	label, err := models.GetGroceryDeliveryLabelByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return label, nil
}

//AddGroceryDeliveryLabel adds a new grocery delivery label
func (r *mutationResolver) AddGroceryDeliveryLabel(ctx context.Context, input models.AddGroceryDeliveryLabelInput) (*models.GroceryDeliveryLabel, error) {
	label := &models.GroceryDeliveryLabel{}
	_ = copier.Copy(&label, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	label.CreatedBy = user.ID
	label, err = models.CreateGroceryDeliveryLabel(*label)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), label.ID.Hex(), "grocery delivery label", label, nil, ctx)
	return label, nil
}

//UpdateGroceryDeliveryLabel updates an existing grocery delivery label
func (r *mutationResolver) UpdateGroceryDeliveryLabel(ctx context.Context, input models.UpdateGroceryDeliveryLabelInput) (*models.GroceryDeliveryLabel, error) {
	label := &models.GroceryDeliveryLabel{}
	label, err := models.GetGroceryDeliveryLabelByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&label, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	label.CreatedBy = user.ID
	label, err = models.UpdateGroceryDeliveryLabel(label)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), label.ID.Hex(), "grocery delivery label", label, nil, ctx)
	return label, nil
}

//DeleteGroceryDeliveryLabel deletes an existing grocery delivery label
func (r *mutationResolver) DeleteGroceryDeliveryLabel(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteGroceryDeliveryLabelByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "grocery delivery label", nil, nil, ctx)
	return &res, err
}

//WineDeliveryLabels gives a list of wine delivery labels
func (r *queryResolver) WineDeliveryLabels(ctx context.Context, wineDeliveryLabelSearch *models.WineDeliveryLabelSearch, text *string, after *string, before *string, first *int, last *int) (*models.WineDeliveryLabelConnection, error) {
	var items []*models.WineDeliveryLabel
	var edges []*models.WineDeliveryLabelEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetWineDeliveryLabels(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.WineDeliveryLabelEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.WineDeliveryLabelConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//WineDeliveryLabel returns a specific wine delivery label by its ID
func (r *queryResolver) WineDeliveryLabel(ctx context.Context, id primitive.ObjectID) (*models.WineDeliveryLabel, error) {
	label, err := models.GetWineDeliveryLabelByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return label, nil
}

//AddWineDeliveryLabel adds a new wine delivery label
func (r *mutationResolver) AddWineDeliveryLabel(ctx context.Context, input models.AddWineDeliveryLabelInput) (*models.WineDeliveryLabel, error) {
	label := &models.WineDeliveryLabel{}
	_ = copier.Copy(&label, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	label.CreatedBy = user.ID
	label, err = models.CreateWineDeliveryLabel(*label)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), label.ID.Hex(), "wine delivery label", label, nil, ctx)
	return label, nil
}

//UpdateWineDeliveryLabel updates an existing wine delivery label
func (r *mutationResolver) UpdateWineDeliveryLabel(ctx context.Context, input models.UpdateWineDeliveryLabelInput) (*models.WineDeliveryLabel, error) {
	label := &models.WineDeliveryLabel{}
	label, err := models.GetWineDeliveryLabelByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&label, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	label.CreatedBy = user.ID
	label, err = models.UpdateWineDeliveryLabel(label)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), label.ID.Hex(), "wine delivery label", label, nil, ctx)
	return label, nil
}

//DeleteWineDeliveryLabel deletes an existing wine delivery label
func (r *mutationResolver) DeleteWineDeliveryLabel(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteWineDeliveryLabelByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "wine delivery label", nil, nil, ctx)
	return &res, err
}

// wineDeliveryLabelResolver is of type struct.
type wineDeliveryLabelResolver struct{ *Resolver }

type foodDeliveryLabelResolver struct{ *Resolver }

type generalLabelResolver struct{ *Resolver }

type groceryDeliveryLabelResolver struct{ *Resolver }
