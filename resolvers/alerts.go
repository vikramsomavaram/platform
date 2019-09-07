/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package resolvers

import (
	"context"
	"encoding/base64"
	"github.com/tribehq/platform/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//DeclineAlertForProviders gives a list of decline alerts for providers
func (r *queryResolver) DeclineAlertForProviders(ctx context.Context, providerType *models.DeclineAlertForProviderType, providerStatus *models.DeclineAlertForProviderStatus, text *string, after *string, before *string, first *int, last *int) (*models.DeclineAlertForProviderConnection, error) {
	var items []*models.DeclineAlertForProvider
	var edges []*models.DeclineAlertForProviderEdge
	filter := bson.D{}
	limit := 25

	items, totalCount, hasPrevious, hasNext, err := models.GetDeclineAlertsForProviders(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.DeclineAlertForProviderEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.DeclineAlertForProviderConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//DeclineAlertForProvider returns a given alert for provider by its ID
func (r *queryResolver) DeclineAlertForProvider(ctx context.Context, id primitive.ObjectID) (*models.DeclineAlertForProvider, error) {
	alert := models.GetDeclineAlertForProviderByID(id.Hex())
	return alert, nil
}

//DeclineAlertForUsers gives a list of decline alerts for users
func (r *queryResolver) DeclineAlertForUsers(ctx context.Context, userType *models.DeclineAlertForUserType, userStatus *models.DeclineAlertForUserStatus, text *string, after *string, before *string, first *int, last *int) (*models.DeclineAlertForUserConnection, error) {
	var items []*models.DeclineAlertForUser
	var edges []*models.DeclineAlertForUserEdge
	filter := bson.D{}
	limit := 25

	items, totalCount, hasPrevious, hasNext, err := models.GetDeclineAlertsForUsers(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.DeclineAlertForUserEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.DeclineAlertForUserConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

//DeclineAlertForUser returns a given alert for users by its ID
func (r *queryResolver) DeclineAlertForUser(ctx context.Context, id primitive.ObjectID) (*models.DeclineAlertForUser, error) {
	alert := models.GetDeclineAlertForUserByID(id.Hex())
	return alert, nil
}

type declineAlertForProviderResolver struct{ *Resolver }

func (declineAlertForProviderResolver) ProviderName(ctx context.Context, obj *models.DeclineAlertForProvider) (string, error) {
	providerID := obj.ProviderName.Hex()
	provider := models.GetServiceProviderByID(providerID)
	return provider.FirstName + " " + provider.LastName, nil
}
