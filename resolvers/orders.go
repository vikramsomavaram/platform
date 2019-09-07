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

//Order returns an order by its ID
func (r *queryResolver) Order(ctx context.Context, id primitive.ObjectID) (*models.Order, error) {
	order, err := models.GetOrderByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return order, nil
}

//Orders gives a list of all orders
func (r *queryResolver) Orders(ctx context.Context, fromDate *time.Time, toDate *time.Time, after *string, before *string, first *int, last *int, orderProviderID *primitive.ObjectID, orderCompanyID *primitive.ObjectID) (*models.OrderConnection, error) {
	var items []*models.Order
	var edges []*models.OrderEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetOrders(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.OrderEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.OrderConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//AddOrder adds a new order
func (r *mutationResolver) AddOrder(ctx context.Context, input models.AddOrderInput) (*models.Order, error) {
	order := &models.Order{}
	_ = copier.Copy(&order, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	order.CreatedBy = user.ID
	order, err = models.CreateOrder(*order)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), order.ID.Hex(), "order", order, nil, ctx)
	return order, nil
}

//UpdateOrder updates an existing order
func (r *mutationResolver) UpdateOrder(ctx context.Context, input models.UpdateOrderInput) (*models.Order, error) {
	order := &models.Order{}
	order, err := models.GetOrderByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&order, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	order.CreatedBy = user.ID
	order, err = models.UpdateOrder(order)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), order.ID.Hex(), "order", order, nil, ctx)
	return order, nil
}

//DeleteOrder deletes an existing order
func (r *mutationResolver) DeleteOrder(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteOrderByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "order", nil, nil, ctx)
	return &res, err
}

//ActivateOrder activates an order by its ID
func (r *mutationResolver) ActivateOrder(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	order, err := models.GetOrderByID(id.Hex())
	if err != nil {
		return nil, err
	}
	order.IsActive = true
	_, err = models.UpdateOrder(order)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "order", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//DeactivateOrder deactivates an order by its ID
func (r *mutationResolver) DeactivateOrder(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	order, err := models.GetOrderByID(id.Hex())
	if err != nil {
		return nil, err
	}
	order.IsActive = false
	_, err = models.UpdateOrder(order)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "order", nil, nil, ctx)
	return utils.PointerBool(false), nil
}

type orderResolver struct{ *Resolver }

func (r *orderResolver) Coupon(ctx context.Context, obj *models.Order) (string, error) {
	return obj.Coupon, nil
}

//PaymentMethod gives the payment method
func (r *orderResolver) PaymentMethod(ctx context.Context, obj *models.Order) (models.PaymentMethodType, error) {
	paymentMethod := models.PaymentMethodType(obj.ID.String())
	return paymentMethod, nil
}

//ExpectedEarning gives expected earning
func (r *orderResolver) ExpectedEarning(ctx context.Context, obj *models.Order) (float64, error) {
	return obj.ExpectedEarning, nil
}

//EarnedAmount gives the earned amount
func (r *orderResolver) EarnedAmount(ctx context.Context, obj *models.Order) (float64, error) {
	return obj.EarnedAmount, nil
}

//CancelOrder cancels orders
func (r *mutationResolver) CancelOrder(ctx context.Context, orderID primitive.ObjectID) (*bool, error) {
	order, err := models.GetOrderByID(orderID.String())
	if err != nil {
		return nil, err
	}
	orderStatus := order.OrderStatus.IsValid()
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Cancelled, user.ID.Hex(), orderID.Hex(), "order", nil, nil, ctx)
	return &orderStatus, nil
}
