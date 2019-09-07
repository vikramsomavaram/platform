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

//Pages gives a list of pages
func (r *queryResolver) Pages(ctx context.Context, searchType *models.PageType, text string, after *string, before *string, first *int, last *int) (*models.PageConnection, error) {
	var items []*models.Page
	var edges []*models.PageEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetPages(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.PageEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.PageConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//Page returns a page by its ID
func (r *queryResolver) Page(ctx context.Context, pageID string) (*models.Page, error) {
	page, err := models.GetPageByID(pageID)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return page, nil
}

//PageDetails returns page details
func (r *queryResolver) PageDetails(ctx context.Context, pageID *string) (*models.Page, error) {
	pageDetails, err := models.GetPageByID(*pageID)
	if err != nil {
		return nil, err
	}
	return pageDetails, nil
}

//AddPage adds a new page
func (r *mutationResolver) AddPage(ctx context.Context, input models.AddPageInput) (*models.Page, error) {
	page := &models.Page{}
	_ = copier.Copy(&page, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	page.CreatedBy = user.ID
	page, err = models.CreatePage(*page)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), page.ID.Hex(), "page", page, nil, ctx)
	return page, nil
}

//UpdatePage updates a page
func (r *mutationResolver) UpdatePage(ctx context.Context, input models.UpdatePageInput) (*models.Page, error) {
	page := &models.Page{}
	page, err := models.GetPageByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&page, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	page.CreatedBy = user.ID
	page, err = models.UpdatePage(page)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), page.ID.Hex(), "page", page, nil, ctx)
	return page, nil
}

//DeletePage deletes a page
func (r *mutationResolver) DeletePage(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeletePageByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "page", nil, nil, ctx)
	return &res, err
}

//DeactivatePage deactivates a page by its ID
func (r *mutationResolver) DeactivatePage(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	page, err := models.GetPageByID(id.Hex())
	if err != nil {
		return nil, err
	}
	page.IsActive = false
	_, err = models.UpdatePage(page)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "page", nil, nil, ctx)
	return utils.PointerBool(false), nil
}

//ActivatePage activates a page by its ID
func (r *mutationResolver) ActivatePage(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	page, err := models.GetPageByID(id.Hex())
	if err != nil {
		return nil, err
	}
	page.IsActive = true
	_, err = models.UpdatePage(page)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "page", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//pageResolver is a type struct.
type pageResolver struct{ *Resolver }
