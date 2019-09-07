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
	"github.com/tribehq/platform/utils/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

//MarketSettings gives a list of market settings
func (r *queryResolver) MarketSettings(ctx context.Context) (*models.MarketSettings, error) {
	return nil, nil
}

//MarketSetting returns a market setting by ID
func (r *queryResolver) MarketSetting(ctx context.Context, marketSettingsID string) (*models.MarketSettings, error) {
	marketSettings, err := models.GetMarketSettingsByID(marketSettingsID)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return marketSettings, nil
}

//SeoSettings gives a list of SEO settings
func (r *queryResolver) SeoSettings(ctx context.Context, seoSettingType *models.SEOSettingType, text *string, after *string, before *string, first *int, last *int) (*models.SEOSettingConnection, error) {
	var items []*models.SEOSetting
	var edges []*models.SEOSettingEdge
	filter := bson.D{}
	limit := 25

	items, totalCount, hasPrevious, hasNext, err := models.GetSEOSettings(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.SEOSettingEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	adBannerList := &models.SEOSettingConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return adBannerList, nil
}

//SeoSetting returns an SEO setting by ID
func (r *queryResolver) SeoSetting(ctx context.Context, id primitive.ObjectID) (*models.SEOSetting, error) {
	seoSetting, err := models.GetSEOSettingByID(id.Hex())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return seoSetting, nil
}

//UpdateSeoSettings updates existing SEO settings
func (r *mutationResolver) UpdateSeoSettings(ctx context.Context, input models.UpdateSEOSettingInput) (*models.SEOSetting, error) {
	seoSettings := &models.SEOSetting{}
	_ = copier.Copy(&seoSettings, &input)
	seoSettings.UpdatedAt = time.Now()
	seoSettings, err := models.CreateSEOSetting(seoSettings)
	if err != nil {
		return nil, err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), seoSettings.ID.Hex(), "seo settings", seoSettings, nil, ctx)
	return seoSettings, nil
}

//UpdateMarketSettings updates existing market settings
func (r *mutationResolver) UpdateMarketSettings(ctx context.Context, input models.UpdateMarketSettingsInput) (*models.MarketSettings, error) {
	marketSettings := &models.MarketSettings{}
	_ = copier.Copy(&marketSettings, &input)
	marketSettings.UpdatedAt = time.Now()
	marketSettings, err := models.CreateMarketSettings(marketSettings)
	if err != nil {
		return nil, err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), marketSettings.ID.Hex(), "market settings", marketSettings, nil, ctx)
	return marketSettings, nil
}

// sEOSettingResolver is of type struct.
type sEOSettingResolver struct{ *Resolver }
