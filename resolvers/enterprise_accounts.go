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

//AddEnterpriseAccount adds a new enterprise account
func (r *mutationResolver) AddEnterpriseAccount(ctx context.Context, input models.AddEnterpriseAccountInput) (*models.EnterpriseAccount, error) {
	enterpriseAccount := &models.EnterpriseAccount{}
	_ = copier.Copy(&enterpriseAccount, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	enterpriseAccount.CreatedBy = user.ID
	enterpriseAccount, err = models.CreateEnterpriseAccount(*enterpriseAccount)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), enterpriseAccount.ID.Hex(), "enterprise account", enterpriseAccount, nil, ctx)
	return enterpriseAccount, nil
}

//UpdateEnterpriseAccount updates an existing enterprise account
func (r *mutationResolver) UpdateEnterpriseAccount(ctx context.Context, input models.UpdateEnterpriseAccountInput) (*models.EnterpriseAccount, error) {
	enterpriseAccount := &models.EnterpriseAccount{}
	enterpriseAccount, err := models.GetEnterpriseAccountByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&enterpriseAccount, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	enterpriseAccount.CreatedBy = user.ID
	enterpriseAccount, err = models.UpdateEnterpriseAccount(enterpriseAccount)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), enterpriseAccount.ID.Hex(), "enterprise account", enterpriseAccount, nil, ctx)
	return enterpriseAccount, nil
}

//DeleteEnterpriseAccount deletes an existing enterprise account
func (r *mutationResolver) DeleteEnterpriseAccount(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteEnterpriseAccountByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "enterprise account", nil, nil, ctx)
	return &res, err
}

//ActivateEnterpriseAccount activates an enterprise account by its ID
func (r *mutationResolver) ActivateEnterpriseAccount(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	enterpriseAccount, err := models.GetEnterpriseAccountByID(id.Hex())
	if err != nil {
		return nil, err
	}
	enterpriseAccount.IsActive = true
	_, err = models.UpdateEnterpriseAccount(enterpriseAccount)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "enterprise account", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//DeactivateEnterpriseAccount deactivates an enterprise account by its ID
func (r *mutationResolver) DeactivateEnterpriseAccount(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	enterpriseAccount, err := models.GetEnterpriseAccountByID(id.Hex())
	if err != nil {
		return nil, err
	}
	enterpriseAccount.IsActive = false
	_, err = models.UpdateEnterpriseAccount(enterpriseAccount)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "enterprise account", nil, nil, ctx)
	return utils.PointerBool(false), nil
}

//EnterpriseAccounts returns a list of enterprise accounts
func (r *queryResolver) EnterpriseAccounts(ctx context.Context, enterpriseAccountType *models.EnterpriseAccountType, enterpriseAccountStatus *models.EnterpriseAccountStatus, text *string, after *string, before *string, first *int, last *int) (*models.EnterpriseAccountConnection, error) {
	var items []*models.EnterpriseAccount
	var edges []*models.EnterpriseAccountEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetEnterpriseAccounts(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.EnterpriseAccountEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.EnterpriseAccountConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//EnterpriseAccount gives an enterprise account by its ID
func (r *queryResolver) EnterpriseAccount(ctx context.Context, id primitive.ObjectID) (*models.EnterpriseAccount, error) {
	enterpriseAccount, err := models.GetEnterpriseAccountByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return enterpriseAccount, nil
}

//EnterpriseAccountPaymentReports gives a list of enterprise account payment reports
func (r *queryResolver) EnterpriseAccountPaymentReports(ctx context.Context, fromDate *time.Time, toDate *time.Time, text *string, after *string, before *string, first *int, last *int) (*models.EnterpriseAccountPaymentReportConnection, error) {
	var items []*models.EnterpriseAccountPaymentReport
	var edges []*models.EnterpriseAccountPaymentReportEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetEnterpriseAccountPaymentReports(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.EnterpriseAccountPaymentReportEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.EnterpriseAccountPaymentReportConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//EnterpriseAccountPaymentReport returns a payment report by its ID
func (r *queryResolver) EnterpriseAccountPaymentReport(ctx context.Context, id primitive.ObjectID) (*models.EnterpriseAccountPaymentReport, error) {
	report, err := models.GetEnterpriseAccountPaymentReportByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return report, nil
}

// enterpriseAccountResolver is of type struct.
type enterpriseAccountResolver struct{ *Resolver }

// enterpriseAccountPaymentReportResolver is of type struct.
type enterpriseAccountPaymentReportResolver struct{ *Resolver }
