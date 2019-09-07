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

//OrderStatusUtilities gives a list of order status utilities
func (r *queryResolver) OrderStatusUtilities(ctx context.Context, orderStatusUtilityType *models.OrderStatusUtilitySearchType, text *string, after *string, before *string, first *int, last *int) (*models.OrderStatusUtilityConnection, error) {
	var items []*models.OrderStatusUtility
	var edges []*models.OrderStatusUtilityEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetOrderstatusUtilities(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.OrderStatusUtilityEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.OrderStatusUtilityConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//OrderStatusUtility returns an order status utility by its ID
func (r *queryResolver) OrderStatusUtility(ctx context.Context, OrderStatusUtilityID primitive.ObjectID) (*models.OrderStatusUtility, error) {
	orderStatusUtility, err := models.GetOrderStatusUtilityByID(OrderStatusUtilityID.String())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return orderStatusUtility, nil
}

//AddOrderStatusUtility adds a new order status utility
func (r *mutationResolver) AddOrderStatusUtility(ctx context.Context, input models.AddOrderStatusUtilityInput) (*models.OrderStatusUtility, error) {
	orderStatusUtility := &models.OrderStatusUtility{}
	_ = copier.Copy(&orderStatusUtility, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	orderStatusUtility.CreatedBy = user.ID
	orderStatusUtility, err = models.CreateOrderStatusUtility(*orderStatusUtility)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), orderStatusUtility.ID.Hex(), "order status utility", orderStatusUtility, nil, ctx)
	return orderStatusUtility, nil
}

//UpdateOrderStatusUtility updates an existing order status utility
func (r *mutationResolver) UpdateOrderStatusUtility(ctx context.Context, input models.UpdateOrderStatusUtilityInput) (*models.OrderStatusUtility, error) {
	orderStatusUtility := &models.OrderStatusUtility{}
	orderStatusUtility, err := models.GetOrderStatusUtilityByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&orderStatusUtility, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	orderStatusUtility.CreatedBy = user.ID
	orderStatusUtility, err = models.UpdateOrderStatusUtility(orderStatusUtility)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), orderStatusUtility.ID.Hex(), "order status utility", orderStatusUtility, nil, ctx)
	return orderStatusUtility, nil
}

//DeleteOrderStatusUtility deletes an existing order status utility
func (r *mutationResolver) DeleteOrderStatusUtility(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteOrderStatusUtilityByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "order status utility", nil, nil, ctx)
	return &res, err
}

//ActivateOrderStatusUtility activates an existing order status utility
func (r *mutationResolver) ActivateOrderStatusUtility(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	orderStatusUtility, err := models.GetOrderStatusUtilityByID(id.Hex())
	if err != nil {
		return nil, err
	}
	orderStatusUtility.IsActive = true
	_, err = models.UpdateOrderStatusUtility(orderStatusUtility)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "order status utility", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//DeactivateOrderStatusUtility deactivates an existing order status utility
func (r *mutationResolver) DeactivateOrderStatusUtility(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	orderStatusUtility, err := models.GetOrderStatusUtilityByID(id.Hex())
	if err != nil {
		return nil, err
	}
	orderStatusUtility.IsActive = false
	_, err = models.UpdateOrderStatusUtility(orderStatusUtility)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "order status utility", nil, nil, ctx)
	return utils.PointerBool(false), nil
}

// orderStatusUtilityResolver is type struct of order status utility.
type orderStatusUtilityResolver struct{ *Resolver }
