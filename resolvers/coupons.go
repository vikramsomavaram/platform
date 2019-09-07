/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package resolvers

import (
	"context"
	"encoding/base64"
	"errors"
	"github.com/jinzhu/copier"
	"github.com/tribehq/platform/lib/audit_log"
	"github.com/tribehq/platform/models"
	"github.com/tribehq/platform/utils"
	"github.com/tribehq/platform/utils/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//Coupons gives a list of coupons
func (r *queryResolver) Coupons(ctx context.Context, couponType *models.CouponType, couponStatus *models.CouponStatus, text *string, after *string, before *string, first *int, last *int) (*models.CouponConnection, error) {
	var items []*models.Coupon
	var edges []*models.CouponEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetCoupons(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.CouponEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := &models.PageInfo{}
	if len(edges) > 1 {
		pageInfo.StartCursor = edges[0].Cursor
		pageInfo.EndCursor = edges[len(edges)-1].Cursor
		pageInfo.HasNextPage = hasNext
		pageInfo.HasPreviousPage = hasPrevious
	}

	itemList := &models.CouponConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//Coupon returns a specific coupon by its ID
func (r *queryResolver) Coupon(ctx context.Context, id primitive.ObjectID) (*models.Coupon, error) {
	coupon := models.GetCouponByID(id.Hex())
	return coupon, nil
}

//AddCoupon adds a new coupon
func (r *mutationResolver) AddCoupon(ctx context.Context, input models.AddCouponInput) (*models.Coupon, error) {
	coupon := &models.Coupon{}
	_ = copier.Copy(&coupon, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	coupon.CreatedBy = user.ID
	coupon, err = models.CreateCoupon(*coupon)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), coupon.ID.Hex(), "coupon", coupon, nil, ctx)
	return coupon, nil

}

//UpdateCoupon updates an existing coupon
func (r *mutationResolver) UpdateCoupon(ctx context.Context, input models.UpdateCouponInput) (*models.Coupon, error) {
	coupon := &models.Coupon{}
	coupon = models.GetCouponByID(input.ID.Hex())
	_ = copier.Copy(&coupon, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	coupon.CreatedBy = user.ID
	coupon, err = models.UpdateCoupon(coupon)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), coupon.ID.Hex(), "coupon", coupon, nil, ctx)
	return coupon, nil
}

//DeleteCoupon deletes an existing coupon
func (r *mutationResolver) DeleteCoupon(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteCouponByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "coupon", nil, nil, ctx)
	return &res, err
}

//DeactivateCoupon deactivates a coupon by its ID
func (r *mutationResolver) DeactivateCoupon(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	coupon := models.GetCouponByID(id.Hex())
	if coupon.ID.IsZero() {
		return utils.PointerBool(false), errors.New("coupon not found")
	}
	coupon.IsActive = false
	_, err := models.UpdateCoupon(coupon)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "coupon", nil, nil, ctx)
	return utils.PointerBool(false), nil

}

//ActivateCoupon activates a coupon by its ID
func (r *mutationResolver) ActivateCoupon(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	coupon := models.GetCouponByID(id.Hex())
	if coupon.ID.IsZero() {
		return utils.PointerBool(false), errors.New("coupon not found")
	}
	coupon.IsActive = true
	_, err := models.UpdateCoupon(coupon)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "coupon", nil, nil, ctx)
	return utils.PointerBool(true), nil

}
