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

//AddCurrency adds a new currency
func (r *mutationResolver) AddCurrency(ctx context.Context, input models.AddCurrencyInput) (*models.Currency, error) {
	currency := &models.Currency{}
	_ = copier.Copy(&currency, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	currency.CreatedBy = user.ID
	currency, err = models.CreateCurrency(*currency)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), currency.ID.Hex(), "currency", currency, nil, ctx)
	return currency, nil
}

//UpdateCurrency updates an existing currency
func (r *mutationResolver) UpdateCurrency(ctx context.Context, input models.UpdateCurrencyInput) (*models.Currency, error) {
	currency := &models.Currency{}
	currency, err := models.GetCurrencyByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&currency, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	currency.CreatedBy = user.ID
	currency, err = models.UpdateCurrency(currency)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), currency.ID.Hex(), "currency", currency, nil, ctx)
	return currency, nil
}

//DeleteCurrency deletes an existing currency
func (r *mutationResolver) DeleteCurrency(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteCurrencyByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "currency", nil, nil, ctx)
	return &res, err
}

//ActivateCurrency activates a currency by its ID
func (r *mutationResolver) ActivateCurrency(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	currency, err := models.GetCurrencyByID(id.Hex())
	if err != nil {
		return nil, err
	}
	currency.IsActive = true
	_, err = models.UpdateCurrency(currency)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "currency", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//DeactivateCurrency deactivates a currency by its ID
func (r *mutationResolver) DeactivateCurrency(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	currency, err := models.GetCurrencyByID(id.Hex())
	if err != nil {
		return nil, err
	}
	currency.IsActive = false
	_, err = models.UpdateCurrency(currency)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "currency", nil, nil, ctx)
	return utils.PointerBool(false), nil
}

//Currency returns a specific currency by its ID
func (r *queryResolver) Currency(ctx context.Context, id primitive.ObjectID) (*models.Currency, error) {
	currency, err := models.GetCurrencyByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return currency, nil
}

//Currencies returns a list of currencies
func (r *queryResolver) Currencies(ctx context.Context, appID primitive.ObjectID, after *string, before *string, first *int, last *int) (*models.CurrencyConnection, error) {
	var items []*models.Currency
	var edges []*models.CurrencyEdge
	filter := bson.D{}
	limit := 25

	items, totalCount, hasPrevious, hasNext, err := models.GetCurrencies(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.CurrencyEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.CurrencyConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}
