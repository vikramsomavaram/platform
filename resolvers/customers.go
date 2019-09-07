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

type customerResolver struct{ *Resolver }

//Customers returns a list of customers
func (r *queryResolver) Customers(ctx context.Context, id primitive.ObjectID, after *string, before *string, first *int, last *int) (*models.CustomerConnection, error) {
	var items []*models.Customer
	var edges []*models.CustomerEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetCustomers(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.CustomerEdge{
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

	itemList := &models.CustomerConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//Customer returns a customer by its ID
func (r *queryResolver) Customer(ctx context.Context, id primitive.ObjectID) (*models.Customer, error) {
	customer, err := models.GetCustomerByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return customer, nil
}

//AddCustomer adds a new customer
func (r *mutationResolver) AddCustomer(ctx context.Context, input models.AddCustomerInput) (*models.Customer, error) {
	customer := &models.Customer{}
	_ = copier.Copy(&customer, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	customer.CreatedBy = user.ID
	customer, err = models.CreateCustomer(*customer)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), customer.ID.Hex(), "customer", customer, nil, ctx)
	return customer, nil
}

//UpdateCustomer updates an existing customer
func (r *mutationResolver) UpdateCustomer(ctx context.Context, input models.UpdateCustomerInput) (*models.Customer, error) {
	customer := &models.Customer{}
	customer, err := models.GetCustomerByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&customer, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	customer.CreatedBy = user.ID
	customer, err = models.UpdateCustomer(customer)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), customer.ID.Hex(), "customer", customer, nil, ctx)
	return customer, nil
}

//DeleteCustomer deletes an existing customer
func (r *mutationResolver) DeleteCustomer(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteCustomerByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "customer", nil, nil, ctx)
	return &res, err
}

//ActivateCustomer activates a customer by their ID
func (r *mutationResolver) ActivateCustomer(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	customer, err := models.GetCustomerByID(id.Hex())
	if err != nil {
		return nil, err
	}
	customer.IsActive = true
	_, err = models.UpdateCustomer(customer)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "customer", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//DeactivateCustomer deactivates a customer by their ID
func (r *mutationResolver) DeactivateCustomer(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	customer, err := models.GetCustomerByID(id.Hex())
	if err != nil {
		return nil, err
	}
	customer.IsActive = false
	_, err = models.UpdateCustomer(customer)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "customer", nil, nil, ctx)
	return utils.PointerBool(false), nil
}
