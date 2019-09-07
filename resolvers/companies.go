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
	"github.com/tribehq/platform/lib/database"
	"github.com/tribehq/platform/models"
	"github.com/tribehq/platform/utils"
	"github.com/tribehq/platform/utils/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

//Companies Resolver
type serviceCompanyResolver struct{ *Resolver }

//Company returns a given company by its ID
func (r *queryResolver) ServiceCompany(ctx context.Context, companyID string) (*models.ServiceCompany, error) {
	company := models.GetServiceCompanyByID(companyID)
	return company, nil
}

//ProvidersCount gives number of workers in a company
func (r *serviceCompanyResolver) ProvidersCount(ctx context.Context, obj *models.ServiceCompany) (int, error) {
	db := database.MongoDB
	filter := bson.D{{"companyId", obj.ID.Hex()}}
	filter = append(filter, bson.E{"deletedAt", bson.M{"$exists": false}}) // we dont want deleted documents
	totalCount, err := db.Collection(models.ServiceProvidersCollection).CountDocuments(context.Background(), filter)
	count := int(totalCount)
	if err != nil {
		return count, err
	}
	return count, nil
}

//DeleteServiceCompany deletes an existing company
func (r *mutationResolver) DeleteServiceCompany(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteServiceCompanyByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "service company", nil, nil, ctx)
	return &res, err
}

//AddServiceCompany adds a new company
func (r *mutationResolver) AddServiceCompany(ctx context.Context, input models.AddServiceCompanyInput) (*models.ServiceCompany, error) {
	company := &models.ServiceCompany{}
	_ = copier.Copy(&company, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	company.CreatedBy = user.ID
	company, err = models.CreateServiceCompany(*company)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), company.ID.Hex(), "service company", company, nil, ctx)
	return company, nil
}

//UpdateServiceCompany updates an existing company
func (r *mutationResolver) UpdateServiceCompany(ctx context.Context, input models.UpdateServiceCompanyInput) (*models.ServiceCompany, error) {
	company := &models.ServiceCompany{}
	company = models.GetServiceCompanyByID(input.ID.Hex())
	_ = copier.Copy(&company, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	company.CreatedBy = user.ID
	company.UpdatedAt = time.Now()
	company, err = models.UpdateServiceCompany(company)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), company.ID.Hex(), "service company", company, nil, ctx)
	return company, nil
}

//ServiceCompanies gives a list of existing companies
func (r *queryResolver) ServiceCompanies(ctx context.Context, companiesType *models.CompaniesSearchType, text *string, companiesStatus *models.CompaniesStatus, after *string, before *string, first *int, last *int) (*models.ServiceCompaniesConnection, error) {
	var items []*models.ServiceCompany
	var edges []*models.ServiceCompanyEdge
	filter := bson.D{}
	limit := 25

	items, totalCount, hasPrevious, hasNext, err := models.GetServiceCompanies(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.ServiceCompanyEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.ServiceCompaniesConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//DeactivateServiceCompany deactivates a company by its ID
func (r *mutationResolver) DeactivateServiceCompany(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	company := models.GetServiceCompanyByID(id.Hex())
	if company.ID.IsZero() {
		return utils.PointerBool(false), errors.New("service company not found")
	}
	company.IsActive = false
	_, err := models.UpdateServiceCompany(company)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "service company", nil, nil, ctx)
	return utils.PointerBool(false), nil
}

//ActivateServiceCompany activates a company by its ID
func (r *mutationResolver) ActivateServiceCompany(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	company := models.GetServiceCompanyByID(id.Hex())
	if company.ID.IsZero() {
		return utils.PointerBool(false), errors.New("service company not found")
	}
	company.IsActive = true
	_, err := models.UpdateServiceCompany(company)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "service company", nil, nil, ctx)
	return utils.PointerBool(true), nil
}
