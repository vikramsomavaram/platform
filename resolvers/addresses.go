/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package resolvers

import (
	"context"
	"encoding/base64"
	"github.com/jinzhu/copier"
	"github.com/tribehq/platform/lib/audit_log"
	"github.com/tribehq/platform/models"
	"github.com/tribehq/platform/utils/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//AddAddress adds address.
func (r *mutationResolver) AddAddress(ctx context.Context, input models.AddAddressInput) (*models.Address, error) {
	address := &models.Address{}
	_ = copier.Copy(&address, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	address.CreatedBy = user.ID
	address, err = models.CreateAddress(*address)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), address.ID.Hex(), "address", address, nil, ctx)
	return address, nil
}

//UpdateAddress updates existing address.
func (r *mutationResolver) UpdateAddress(ctx context.Context, input models.UpdateAddressInput) (*models.Address, error) {
	address := &models.Address{}
	address = models.GetAddressByID(input.ID.Hex())
	_ = copier.Copy(&address, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	address.CreatedBy = user.ID
	address, err = models.UpdateAddress(address)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), address.ID.Hex(), "address", address, nil, ctx)
	return address, nil
}

//DeleteAddress deletes address.
func (r *mutationResolver) DeleteAddress(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteAddressByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "address", nil, nil, ctx)
	return &res, err
}

//SavedAddresses returns list of saved address.
func (r *queryResolver) SavedAddresses(ctx context.Context, id primitive.ObjectID, after *string, before *string, first *int, last *int) (*models.AddressConnection, error) {
	var items []*models.Address
	var edges []*models.AddressEdge
	filter := bson.D{}
	limit := 25

	items, totalCount, hasPrevious, hasNext, err := models.GetAddresses(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.AddressEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.AddressConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//SavedAddress returns saved address by id.
func (r *queryResolver) SavedAddress(ctx context.Context, id primitive.ObjectID) (*models.Address, error) {
	address := models.GetAddressByID(id.Hex())
	return address, nil
}
