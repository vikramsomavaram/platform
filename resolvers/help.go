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

//AddHelpDetail adds a new help detail
func (r *mutationResolver) AddHelpDetail(ctx context.Context, input models.AddHelpDetailInput) (*models.HelpDetail, error) {
	help := &models.HelpDetail{}
	_ = copier.Copy(&help, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	help.CreatedBy = user.ID
	help, err = models.CreateHelpDetail(*help)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), help.ID.Hex(), "help", help, nil, ctx)
	return help, nil
}

//UpdateHelpDetail updates an existing help detail
func (r *mutationResolver) UpdateHelpDetail(ctx context.Context, input models.UpdateHelpDetailInput) (*models.HelpDetail, error) {
	help := &models.HelpDetail{}
	help, err := models.GetHelpDetailByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&help, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	help.CreatedBy = user.ID
	help, err = models.UpdateHelpDetail(help)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), help.ID.Hex(), "help", help, nil, ctx)
	return help, nil
}

//DeleteHelpDetail deletes an existing help detail
func (r *mutationResolver) DeleteHelpDetail(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteHelpDetailByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "help", nil, nil, ctx)
	return &res, err
}

//ActivateHelpDetail activates a help detail by its ID
func (r *mutationResolver) ActivateHelpDetail(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	help, err := models.GetHelpDetailByID(id.Hex())
	if err != nil {
		return nil, err
	}
	help.IsActive = true
	_, err = models.UpdateHelpDetail(help)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "help", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//DeactivateHelpDetail deactivates a help detail by its ID
func (r *mutationResolver) DeactivateHelpDetail(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	help, err := models.GetHelpDetailByID(id.Hex())
	if err != nil {
		return nil, err
	}
	help.IsActive = false
	_, err = models.UpdateHelpDetail(help)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "help", nil, nil, ctx)
	return utils.PointerBool(false), nil
}

//AddHelpCategory adds a new help category
func (r *mutationResolver) AddHelpCategory(ctx context.Context, input models.AddHelpCategoryInput) (*models.HelpCategory, error) {
	helpCategory := &models.HelpCategory{}
	_ = copier.Copy(&helpCategory, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	helpCategory.CreatedBy = user.ID
	helpCategory, err = models.CreateHelpCategory(*helpCategory)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), helpCategory.ID.Hex(), "help category", helpCategory, nil, ctx)
	return helpCategory, nil
}

//UpdateHelpCategory updates an existing help category
func (r *mutationResolver) UpdateHelpCategory(ctx context.Context, input models.UpdateHelpCategoryInput) (*models.HelpCategory, error) {
	helpCategory := &models.HelpCategory{}
	helpCategory, err := models.GetHelpCategoryByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&helpCategory, &input)
	id, err := primitive.ObjectIDFromHex(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	helpCategory.CreatedBy = user.ID
	helpCategory.ID = id
	helpCategory.UpdatedAt = time.Now()
	helpCategory, err = models.UpdateHelpCategory(helpCategory)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), helpCategory.ID.Hex(), "help category", helpCategory, nil, ctx)
	return helpCategory, nil
}

//DeleteHelpCategory deletes an existing help category
func (r *mutationResolver) DeleteHelpCategory(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteHelpCategoryByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "help category", nil, nil, ctx)
	return &res, err
}

//HelpDetails gives a list of help details
func (r *queryResolver) HelpDetails(ctx context.Context, helpDetailType *models.HelpDetailType, text *string, after *string, before *string, first *int, last *int) (*models.HelpDetailConnection, error) {
	var items []*models.HelpDetail
	var edges []*models.HelpDetailEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetHelpDetails(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.HelpDetailEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.HelpDetailConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//HelpDetail returns a help detail by its ID
func (r *queryResolver) HelpDetail(ctx context.Context, id primitive.ObjectID) (*models.HelpDetail, error) {
	help, err := models.GetHelpDetailByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return help, nil
}

//HelpCategories gives a list of help categories
func (r *queryResolver) HelpCategories(ctx context.Context, helpCategoryType *models.HelpCategoryType, text *string, after *string, before *string, first *int, last *int) (*models.HelpCategoryConnection, error) {
	var items []*models.HelpCategory
	var edges []*models.HelpCategoryEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetHelpCategories(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.HelpCategoryEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.HelpCategoryConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//HelpCategory returns a help category by its ID
func (r *queryResolver) HelpCategory(ctx context.Context, id primitive.ObjectID) (*models.HelpCategory, error) {
	helpCategory, err := models.GetHelpCategoryByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return helpCategory, nil
}

//ActivateHelpCategory activates a help category by its ID
func (r *mutationResolver) ActivateHelpCategory(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	helpCategory, err := models.GetHelpCategoryByID(id.Hex())
	if err != nil {
		return nil, err
	}
	helpCategory.IsActive = true
	_, err = models.UpdateHelpCategory(helpCategory)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "help category", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//DeactivateHelpCategory deactivates a help category by its ID
func (r *mutationResolver) DeactivateHelpCategory(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	helpCategory, err := models.GetHelpCategoryByID(id.Hex())
	if err != nil {
		return nil, err
	}
	helpCategory.IsActive = false
	_, err = models.UpdateHelpCategory(helpCategory)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "help category", nil, nil, ctx)
	return utils.PointerBool(false), nil
}

type helpCategoryResolver struct{ *Resolver }

type helpDetailResolver struct{ *Resolver }
