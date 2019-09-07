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
	"github.com/tribehq/platform/models"
	"github.com/tribehq/platform/utils"
	"github.com/tribehq/platform/utils/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//AddAdvertisementBanner adds a new advertisement banner
func (r *mutationResolver) AddAdvertisementBanner(ctx context.Context, input models.AddBannerInput) (*models.AdvertisementBanner, error) {
	adBanner := &models.AdvertisementBanner{}
	_ = copier.Copy(&adBanner, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	adBanner.CreatedBy = user.ID
	adBanner, err = models.CreateAdvertisementBanner(*adBanner)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), adBanner.ID.Hex(), "advertisement banner", adBanner, nil, ctx)
	return adBanner, nil
}

//AdvertisementBanners gives a list of advertisement banners
func (r *queryResolver) AdvertisementBanners(ctx context.Context, bannersType *models.BannersType, text *string, bannerStatus *models.BannerStatusInput, after *string, before *string, first *int, last *int) (*models.AdvertisementBannerConnection, error) {
	var items []*models.AdvertisementBanner
	var edges []*models.AdvertisementBannerEdge
	filter := bson.D{}
	limit := 25

	items, totalCount, hasPrevious, hasNext, err := models.GetAdvertisementBanners(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.AdvertisementBannerEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	adBannerList := &models.AdvertisementBannerConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return adBannerList, nil
}

//AdvertisementBanner returns a given banner by its ID
func (r *queryResolver) AdvertisementBanner(ctx context.Context, id primitive.ObjectID) (*models.AdvertisementBanner, error) {
	adBanner := models.GetAdvertisementBannerByID(id.Hex())
	return adBanner, nil
}

//UpdateAdvertisementBanner updates an advertisement banner
func (r *mutationResolver) UpdateAdvertisementBanner(ctx context.Context, input models.UpdateBannerInput) (*models.AdvertisementBanner, error) {
	adBanner := &models.AdvertisementBanner{}
	adBanner = models.GetAdvertisementBannerByID(input.ID.Hex())
	_ = copier.Copy(&adBanner, &input)

	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	adBanner.CreatedBy = user.ID
	adBanner, err = models.UpdateAdvertisementBanner(adBanner)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), adBanner.ID.Hex(), "advertisement banner", adBanner, nil, ctx)
	return adBanner, nil
}

//DeleteAdvertisementBanner deletes an advertisement banner
func (r *mutationResolver) DeleteAdvertisementBanner(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	res, err := models.DeleteAdvertisementBannerByID(id.Hex())
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "advertisement banner", nil, nil, ctx)
	return &res, err
}

//ActivateAdvertisementBanner activates an advertisement banner by given ID
func (r *mutationResolver) ActivateAdvertisementBanner(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	adBanner := models.GetAdvertisementBannerByID(id.Hex())
	if adBanner.ID.IsZero() {
		return utils.PointerBool(false), errors.New("advertisement banner not found")
	}
	adBanner.IsActive = true
	_, err := models.UpdateAdvertisementBanner(adBanner)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "advertisement banner", nil, nil, ctx)
	return utils.PointerBool(true), nil
}

//DeactivateAdvertisementBanner deactivates an advertisement banner by given ID
func (r *mutationResolver) DeactivateAdvertisementBanner(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	adBanner := models.GetAdvertisementBannerByID(id.Hex())
	if adBanner.ID.IsZero() {
		return utils.PointerBool(false), errors.New("advertisement banner not found")
	}
	adBanner.IsActive = false
	_, err := models.UpdateAdvertisementBanner(adBanner)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "advertisement banner", nil, nil, ctx)
	return utils.PointerBool(false), nil
}
