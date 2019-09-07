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

//AddFAQ adds a new FAQ
func (r *mutationResolver) AddFaq(ctx context.Context, input models.AddFAQInput) (*models.FAQ, error) {
	faq := &models.FAQ{}
	_ = copier.Copy(&faq, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	faq.CreatedBy = user.ID
	categoryID, err := primitive.ObjectIDFromHex(input.Category)
	faq.Category = categoryID
	faq, err = models.CreateFAQ(*faq)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), faq.ID.Hex(), "faq", faq, nil, ctx)
	return faq, nil
}

//UpdateFAQ updates an existing FAQ
func (r *mutationResolver) UpdateFaq(ctx context.Context, input models.UpdateFAQInput) (*models.FAQ, error) {
	faq := &models.FAQ{}
	faq, err := models.GetFAQByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&faq, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	faq.CreatedBy = user.ID
	categoryID, err := primitive.ObjectIDFromHex(input.Category)
	faq.Category = categoryID
	faq, err = models.UpdateFAQ(faq)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), faq.ID.Hex(), "faq", faq, nil, ctx)
	return faq, nil
}

//DeleteFAQ deletes an existing FAQ
func (r *mutationResolver) DeleteFaq(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteFAQByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "faq", nil, nil, ctx)
	return &res, err
}

//ActivateFAQ activates a FAQ by its ID
func (r *mutationResolver) ActivateFaq(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	faq, err := models.GetFAQByID(id.Hex())
	if err != nil {
		return nil, err
	}
	faq.IsActive = true
	_, err = models.UpdateFAQ(faq)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "faq", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//DeactivateFAQ deactivates a FAQ by its ID
func (r *mutationResolver) DeactivateFaq(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	faq, err := models.GetFAQByID(id.Hex())
	if err != nil {
		return nil, err
	}
	faq.IsActive = false
	_, err = models.UpdateFAQ(faq)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "faq", nil, nil, ctx)
	return utils.PointerBool(false), nil
}

//FAQs gives a list of FAQs
func (r *queryResolver) Faqs(ctx context.Context, faqType *models.FAQType, text *string, after *string, before *string, first *int, last *int) (*models.FAQConnection, error) {
	var items []*models.FAQ
	var edges []*models.FAQEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetFAQs(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.FAQEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.FAQConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//FAQ returns a specific FAQ by its ID
func (r *queryResolver) Faq(ctx context.Context, id primitive.ObjectID) (*models.FAQ, error) {
	faq, err := models.GetFAQByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return faq, nil
}

//FAQCategories gives a list of FAQ categories
func (r *queryResolver) FaqCategories(ctx context.Context, faqCategoryType *models.FAQCategorySearchType, text *string, after *string, before *string, first *int, last *int) (*models.FAQCategoryConnection, error) {
	var items []*models.FAQCategory
	var edges []*models.FAQCategoryEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetFAQCategories(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.FAQCategoryEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.FAQCategoryConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//FAQCategory returns a specific FAQ category by its ID
func (r *queryResolver) FaqCategory(ctx context.Context, id primitive.ObjectID) (*models.FAQCategory, error) {
	faqCategory, err := models.GetFAQCategoryByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return faqCategory, nil
}

//AddFAQCategory adds a new FAQ category
func (r *mutationResolver) AddFAQCategory(ctx context.Context, input models.AddFAQCategoryInput) (*models.FAQCategory, error) {
	faqCategory := &models.FAQCategory{}
	_ = copier.Copy(&faqCategory, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	faqCategory.CreatedBy = user.ID
	faqCategory, err = models.CreateFAQCategory(*faqCategory)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), faqCategory.ID.Hex(), "faq category", faqCategory, nil, ctx)
	return faqCategory, nil
}

//UpdateFAQCategory updates an existing FAQ category
func (r *mutationResolver) UpdateFAQCategory(ctx context.Context, input models.UpdateFAQCategoryInput) (*models.FAQCategory, error) {
	faqCategory := &models.FAQCategory{}
	faqCategory, err := models.GetFAQCategoryByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&faqCategory, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	faqCategory.CreatedBy = user.ID
	faqCategory.UpdatedAt = time.Now()
	faqCategory, err = models.UpdateFAQCategory(faqCategory)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), faqCategory.ID.Hex(), "faq category", faqCategory, nil, ctx)
	return faqCategory, nil
}

//DeleteFAQCategory deletes an existing FAQ category
func (r *mutationResolver) DeleteFAQCategory(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteFAQCategoryByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "faq category", nil, nil, ctx)
	return &res, err
}

//ActivateFAQCategory activates a FAQ category by its ID
func (r *mutationResolver) ActivateFAQCategory(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	faqCategory, err := models.GetFAQCategoryByID(id.Hex())
	if err != nil {
		return nil, err
	}
	faqCategory.IsActive = true
	_, err = models.UpdateFAQCategory(faqCategory)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "faq category", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//DeactivateFAQCategory deactivates a FAQ category by its ID
func (r *mutationResolver) DeactivateFAQCategory(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	faqCategory, err := models.GetFAQCategoryByID(id.Hex())
	if err != nil {
		return nil, err
	}
	faqCategory.IsActive = false
	_, err = models.UpdateFAQCategory(faqCategory)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "faq category", nil, nil, ctx)
	return utils.PointerBool(false), nil
}

// fAQResolver is of type struct.
type fAQResolver struct{ *Resolver }

//Category returns faq category.
func (fAQResolver) Category(ctx context.Context, obj *models.FAQ) (*models.FAQCategory, error) {
	category, err := models.GetFAQCategoryByID(obj.Category.Hex())
	if err != nil {
		return nil, err
	}
	return category, nil
}

// fAQCategoryResolver is of type struct.
type fAQCategoryResolver struct{ *Resolver }
