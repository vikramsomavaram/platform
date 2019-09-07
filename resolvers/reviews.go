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

//StoreReviews gives a list of store reviews
func (r *queryResolver) StoreReviews(ctx context.Context, storeReviewType *models.StoreReviewType, text *string, after *string, before *string, first *int, last *int) (*models.StoreReviewConnection, error) {
	var items []*models.StoreReview
	var edges []*models.StoreReviewEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetStoreReviews(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.StoreReviewEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.StoreReviewConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//StoreReview returns a store review by ID
func (r *queryResolver) StoreReview(ctx context.Context, id primitive.ObjectID) (*models.StoreReview, error) {
	storeOrderReview, err := models.GetStoreReviewByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return storeOrderReview, nil
}

//AddStoreReview adds a new review
func (r *mutationResolver) AddStoreReview(ctx context.Context, input models.AddReviewInput) (*models.StoreReview, error) {
	storeReview := &models.StoreReview{}
	_ = copier.Copy(&storeReview, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	storeReview.CreatedBy = user.ID
	storeReview, err = models.CreateStoreReview(*storeReview)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), storeReview.ID.Hex(), "store review", storeReview, nil, ctx)
	return storeReview, nil
}

//UpdateStoreReview updates an existing review
func (r *mutationResolver) UpdateStoreReview(ctx context.Context, input models.UpdateReviewInput) (*models.StoreReview, error) {
	storeReview := &models.StoreReview{}
	_ = copier.Copy(&storeReview, &input)
	storeReview.UpdatedAt = time.Now()
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	storeReview, err = models.UpdateStoreReview(storeReview)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), storeReview.ID.Hex(), "store review", storeReview, nil, ctx)
	return storeReview, nil
}

//DeleteStoreOrderReview deletes an existing store order review
func (r *mutationResolver) DeleteStoreReview(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteStoreReviewByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "store review", nil, nil, ctx)
	return &res, err
}

//ActivateStoreReview activates a store order review by ID
func (r *mutationResolver) ActivateStoreReview(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	storeOrderReview, err := models.GetStoreReviewByID(id.Hex())
	if err != nil {
		return nil, err
	}
	storeOrderReview.IsActive = true
	_, err = models.UpdateStoreReview(storeOrderReview)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "store review", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//DeactivateStoreReview deactivates a store order review by ID
func (r *mutationResolver) DeactivateStoreReview(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	storeOrderReview, err := models.GetStoreReviewByID(id.Hex())
	if err != nil {
		return nil, err
	}
	storeOrderReview.IsActive = false
	_, err = models.UpdateStoreReview(storeOrderReview)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "store review", nil, nil, ctx)
	return utils.PointerBool(false), nil
}

//Review returns a review by its ID
func (r *queryResolver) Review(ctx context.Context, id primitive.ObjectID) (*models.Review, error) {
	review, err := models.GetReviewByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return review, nil
}

//Reviews gives a list of reviews
func (r *queryResolver) Reviews(ctx context.Context, reviewType *models.ReviewType, text *string, after *string, before *string, first *int, last *int) (*models.ReviewConnection, error) {
	var items []*models.Review
	var edges []*models.ReviewEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetReviews(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.ReviewEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.ReviewConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//AddReview adds a new review
func (r *mutationResolver) AddReview(ctx context.Context, input models.AddReviewInput) (*models.Review, error) {
	review := &models.Review{}
	_ = copier.Copy(&review, &input)
	review, err := models.CreateReview(*review)
	if err != nil {
		return nil, err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	review.CreatedBy = user.ID
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), review.ID.Hex(), "review", review, nil, ctx)
	return review, nil
}

//UpdateReview updates an existing review
func (r *mutationResolver) UpdateReview(ctx context.Context, input models.UpdateReviewInput) (*models.Review, error) {
	review := &models.Review{}
	_ = copier.Copy(&review, &input)
	review.UpdatedAt = time.Now()
	review, err := models.UpdateReview(review)
	if err != nil {
		return nil, err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	review.CreatedBy = user.ID
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), review.ID.Hex(), "review", review, nil, ctx)
	return review, nil
}

//DeleteReview deletes an existing review
func (r *mutationResolver) DeleteReview(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteReviewByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "review", nil, nil, ctx)
	return &res, err
}

//ActivateReview activates a store order review by ID
func (r *mutationResolver) ActivateReview(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	review, err := models.GetReviewByID(id.Hex())
	if err != nil {
		return nil, err
	}
	review.IsActive = true
	_, err = models.UpdateReview(review)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "review", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//DeactivateReview deactivates a store order review by ID
func (r *mutationResolver) DeactivateReview(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	review, err := models.GetReviewByID(id.Hex())
	if err != nil {
		return nil, err
	}
	review.IsActive = false
	_, err = models.UpdateReview(review)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "review", nil, nil, ctx)
	return utils.PointerBool(false), nil
}

// reviewResolver is of type struct.
type reviewResolver struct{ *Resolver }

func (reviewResolver) Date(ctx context.Context, obj *models.Review) (*time.Time, error) {
	panic("implement me")
}

func (reviewResolver) IsActive(ctx context.Context, obj *models.Review) (string, error) {
	panic("implement me")
}

//UserName gives user name
func (reviewResolver) UserName(ctx context.Context, obj *models.Review) (string, error) {
	user := models.GetUserByID(obj.UserID)
	userName := user.FirstName + " " + user.LastName
	return userName, nil
}

//ProviderName gives provider name
func (reviewResolver) ProviderName(ctx context.Context, obj *models.Review) (string, error) {
	provider := models.GetServiceProviderByID(obj.ProviderID)
	providerName := provider.FirstName + " " + provider.LastName
	return providerName, nil
}
