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

type requiredDocumentResolver struct{ *Resolver }

//CompanyProviderDocuments gives a list of company provider documents
func (r *queryResolver) CompanyProviderDocuments(ctx context.Context, companyID *string, after *string, before *string, first *int, last *int) (*models.DocumentConnection, error) {
	var items []*models.Document
	var edges []*models.DocumentEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetDocuments(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.DocumentEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.DocumentConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//CompanyProviderDocument returns a specific company provider document by its ID
func (r *queryResolver) CompanyProviderDocument(ctx context.Context, id primitive.ObjectID) (*models.Document, error) {
	document, err := models.GetDocumentByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return document, nil
}

//ServiceProviderDocuments gives a list of document provider documents
func (r *queryResolver) ServiceProviderDocuments(ctx context.Context, documentProviderID *string, after *string, before *string, first *int, last *int) (*models.DocumentConnection, error) {
	var items []*models.Document
	var edges []*models.DocumentEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetDocuments(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.DocumentEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.DocumentConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//ServiceProviderDocument returns a specific document provider document by its ID
func (r *queryResolver) ServiceProviderDocument(ctx context.Context, id primitive.ObjectID) (*models.Document, error) {
	document, err := models.GetDocumentByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return document, nil
}

type documentResolver struct{ *Resolver }

//CreatedAt gives the date and time of creation
func (r *documentResolver) CreatedAt(ctx context.Context, obj *models.Document) (string, error) {
	return obj.CreatedAt.String(), nil
}

//UpdatedAt gives the date and time of updation
func (r *documentResolver) UpdatedAt(ctx context.Context, obj *models.Document) (string, error) {
	return obj.UpdatedAt.String(), nil
}

//ExpiryDate gives date of expiry
func (r *documentResolver) ExpiryDate(ctx context.Context, obj *models.Document) (string, error) {
	return obj.ExpiryDate.String(), nil
}

//AddDocument adds a new document
func (r *mutationResolver) AddDocument(ctx context.Context, input models.AddDocumentInput) (*models.Document, error) {
	document := &models.Document{}
	_ = copier.Copy(&document, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	document.CreatedBy = user.ID
	document, err = models.CreateDocument(*document)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), document.ID.Hex(), "document", document, nil, ctx)
	return document, nil
}

//UpdateDocument updates a new document
func (r *mutationResolver) UpdateDocument(ctx context.Context, input models.UpdateDocumentInput) (*models.Document, error) {
	document := &models.Document{}
	document, err := models.GetDocumentByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&document, &input)
	id, err := primitive.ObjectIDFromHex(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	document.CreatedBy = user.ID
	document.ID = id
	document, err = models.UpdateDocument(document)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), document.ID.Hex(), "document", document, nil, ctx)
	return document, nil
}

//DeleteDocument deletes an existing document
func (r *mutationResolver) DeleteDocument(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteDocumentByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "document", nil, nil, ctx)
	return &res, err
}

//ActivateDocument activates a document by its ID
func (r *mutationResolver) ActivateDocument(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	document, err := models.GetDocumentByID(id.Hex())
	if err != nil {
		return nil, err
	}
	document.IsActive = true
	_, err = models.UpdateDocument(document)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "document", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//DeactivateDocument deactivate document by its ID
func (r *mutationResolver) DeactivateDocument(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	document, err := models.GetDocumentByID(id.Hex())
	if err != nil {
		return nil, err
	}
	document.IsActive = false
	_, err = models.UpdateDocument(document)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "document", nil, nil, ctx)
	return utils.PointerBool(false), nil
}

//AddRequiredDocument adds a new required document
func (r *mutationResolver) AddRequiredDocument(ctx context.Context, input models.AddManageDocumentInput) (*models.RequiredDocument, error) {
	requiredDocument := &models.RequiredDocument{}
	_ = copier.Copy(&requiredDocument, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	requiredDocument.CreatedBy = user.ID
	requiredDocument, err = models.CreateRequiredDocument(*requiredDocument)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), requiredDocument.ID.Hex(), "required document", requiredDocument, nil, ctx)
	return requiredDocument, nil
}

//UpdateRequiredDocument updates an existing required document
func (r *mutationResolver) UpdateRequiredDocument(ctx context.Context, input models.UpdateManageDocumentInput) (*models.RequiredDocument, error) {
	requiredDocument := &models.RequiredDocument{}
	requiredDocument, err := models.GetRequiredDocumentByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&requiredDocument, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	requiredDocument.CreatedBy = user.ID
	requiredDocument, err = models.UpdateRequiredDocument(requiredDocument)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), requiredDocument.ID.Hex(), "required document", requiredDocument, nil, ctx)
	return requiredDocument, nil
}

//DeleteRequiredDocument deletes an existing required document
func (r *mutationResolver) DeleteRequiredDocument(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteRequiredDocumentByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "required document", nil, nil, ctx)
	return &res, err
}

//ActivateRequiredDocument activates a required document by its ID
func (r *mutationResolver) ActivateRequiredDocument(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	requiredDocument, err := models.GetRequiredDocumentByID(id.Hex())
	if err != nil {
		return nil, err
	}
	requiredDocument.IsActive = true
	_, err = models.UpdateRequiredDocument(requiredDocument)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "required document", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//DeactivateRequiredDocument activates a required document by its ID
func (r *mutationResolver) DeactivateRequiredDocument(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	requiredDocument, err := models.GetRequiredDocumentByID(id.Hex())
	if err != nil {
		return nil, err
	}
	requiredDocument.IsActive = false
	_, err = models.UpdateRequiredDocument(requiredDocument)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "required document", nil, nil, ctx)
	return utils.PointerBool(false), nil
}

func (r *queryResolver) RequiredDocuments(ctx context.Context, id *string, after *string, before *string, first *int, last *int) (*models.RequiredDocumentConnection, error) {
	var items []*models.RequiredDocument
	var edges []*models.RequiredDocumentEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetRequiredDocuments(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.RequiredDocumentEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.RequiredDocumentConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

func (r *queryResolver) RequiredDocument(ctx context.Context, id primitive.ObjectID) (*models.RequiredDocument, error) {
	requiredDocument, err := models.GetRequiredDocumentByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return requiredDocument, nil
}
