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

type oAuthApplicationResolver struct{ *Resolver }

func (oAuthApplicationResolver) CreatedBy(ctx context.Context, obj *models.OAuthApplication) (string, error) {
	return obj.CreatedBy.Hex(), nil
}

//AddOAuthApplication adds a new OAuth application
func (r *mutationResolver) AddOAuthApplication(ctx context.Context, input *models.AddOAuthApplicationInput) (*models.OAuthApplication, error) {
	oAuthApplication := &models.OAuthApplication{}
	_ = copier.Copy(&oAuthApplication, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	oAuthApplication.ClientID = utils.RandomIDGen(16)
	oAuthApplication.ClientSecret = utils.RandomIDGen(32)
	oAuthApplication.AccessTokenValiditySecs = 3600
	oAuthApplication.CreatedBy = user.ID
	oAuthApplication, err = models.CreateOAuthApplication(*oAuthApplication)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), oAuthApplication.ID.Hex(), "oauth application", oAuthApplication, nil, ctx)
	return oAuthApplication, nil
}

//UpdateOAuthApplication updates an existing OAuth application
func (r *mutationResolver) UpdateOAuthApplication(ctx context.Context, input *models.UpdateOAuthApplicationInput) (*models.OAuthApplication, error) {
	oAuthApplication := &models.OAuthApplication{}
	oAuthApplication, err := models.GetOAuthApplicationByID(input.ID.Hex())
	if err != nil {
		return nil, err
	}
	_ = copier.Copy(&oAuthApplication, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	oAuthApplication.CreatedBy = user.ID
	oAuthApplication, err = models.UpdateOAuthApplication(oAuthApplication)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), oAuthApplication.ID.Hex(), "oauth application", oAuthApplication, nil, ctx)
	return oAuthApplication, nil

}

//DeleteOAuthApplication deletes an existing OAuth application
func (r *mutationResolver) DeleteOAuthApplication(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteOAuthApplicationByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "oauth application", nil, nil, ctx)
	return &res, err
}

//ActivateOAuthApplication activates an OAuth application by its ID
func (r *mutationResolver) ActivateOAuthApplication(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	oAuthApplication, err := models.GetOAuthApplicationByID(id.Hex())
	if err != nil {
		return nil, err
	}
	oAuthApplication.IsActive = true
	_, err = models.UpdateOAuthApplication(oAuthApplication)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "oauth application", nil, nil, ctx)
	return utils.PointerBool(true), nil

}

//DeactivateOAuthApplication deactivates an OAuth application by its ID
func (r *mutationResolver) DeactivateOAuthApplication(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	oAuthApplication, err := models.GetOAuthApplicationByID(id.Hex())
	if err != nil {
		return nil, err
	}
	oAuthApplication.IsActive = false
	_, err = models.UpdateOAuthApplication(oAuthApplication)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "oauth application", nil, nil, ctx)
	return utils.PointerBool(false), nil

}

//OAuthApplications gives a list of OAuth applications
func (r *queryResolver) OAuthApplications(ctx context.Context, appStatus *models.AppStatus, text *string, after *string, before *string, first *int, last *int) (*models.OAuthApplicationConnection, error) {
	var items []*models.OAuthApplication
	var edges []*models.OAuthApplicationEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetOAuthApplications(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.OAuthApplicationEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.OAuthApplicationConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil

}

//OAuthApplication returns an OAuth application by its ID
func (r *queryResolver) OAuthApplication(ctx context.Context, appID primitive.ObjectID) (*models.OAuthApplication, error) {
	oAuthApplication, err := models.GetOAuthApplicationByID(appID.String())
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return oAuthApplication, nil
}

//TODO
//OAuthApplicationStatistics gives OAuth statistics
func (r *queryResolver) OAuthApplicationStatistics(ctx context.Context, appID primitive.ObjectID) (*models.OAuthApplicationStatistics, error) {
	panic("implement me")
}

//TODO
//SubmitAppForApproval submits app for approval
func (r *mutationResolver) SubmitAppForApproval(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	panic("implement me")
}

//TODO
//RevokeAccessToken revokes access token
func (r *mutationResolver) RevokeAccessToken(ctx context.Context, token string) (*bool, error) {
	panic("implement me")
}

//TODO
//ResetApplicationClientSecret changes client secret
func (r *mutationResolver) ResetApplicationClientSecret(ctx context.Context, appID primitive.ObjectID) (*string, error) {
	panic("implement me")
}
