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

type orderNoteResolver struct{ *Resolver }

//OrderNotes give a list of order notes
func (r *queryResolver) OrderNotes(ctx context.Context, id primitive.ObjectID, after *string, before *string, first *int, last *int) (*models.OrderNoteConnection, error) {
	var items []*models.OrderNote
	var edges []*models.OrderNoteEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetOrderNotes(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.OrderNoteEdge{
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

	itemList := &models.OrderNoteConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//OrderNote return an order noter by ID
func (r *queryResolver) OrderNote(ctx context.Context, id primitive.ObjectID) (*models.OrderNote, error) {
	orderNote, err := models.GetOrderNoteByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return orderNote, nil
}

//AddOrderNote adds a new order note
func (r *mutationResolver) AddOrderNote(ctx context.Context, input models.AddOrderNoteInput) (*models.OrderNote, error) {
	orderNote := &models.OrderNote{}
	_ = copier.Copy(&orderNote, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	orderNote.CreatedBy = user.ID
	orderNote, err = models.CreateOrderNote(*orderNote)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), orderNote.ID.Hex(), "order note", orderNote, nil, ctx)
	return orderNote, nil
}

//UpdateOrderNote updates an existing order note
func (r *mutationResolver) UpdateOrderNote(ctx context.Context, input models.UpdateOrderNoteInput) (*models.OrderNote, error) {
	orderNote := &models.OrderNote{}
	orderNote, err := models.GetOrderNoteByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&orderNote, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	orderNote.CreatedBy = user.ID
	orderNote, err = models.UpdateOrderNote(orderNote)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), orderNote.ID.Hex(), "order note", orderNote, nil, ctx)
	return orderNote, nil
}

//DeleteOrderNote deletes an existing order note
func (r *mutationResolver) DeleteOrderNote(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteOrderNoteByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "order note", nil, nil, ctx)
	return &res, err
}

//ActivateOrderNote activates an order note by its ID
func (r *mutationResolver) ActivateOrderNote(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	orderNote, err := models.GetOrderNoteByID(id.Hex())
	if err != nil {
		return nil, err
	}
	orderNote.IsActive = true
	_, err = models.UpdateOrderNote(orderNote)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "order note", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//DeactivateOrderNote deactivates an order note by its ID
func (r *mutationResolver) DeactivateOrderNote(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	orderNote, err := models.GetOrderNoteByID(id.Hex())
	if err != nil {
		return nil, err
	}
	orderNote.IsActive = false
	_, err = models.UpdateOrderNote(orderNote)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "order note", nil, nil, ctx)
	return utils.PointerBool(false), nil
}
