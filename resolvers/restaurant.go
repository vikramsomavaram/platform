/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package resolvers

import (
	"context"
	"encoding/base64"
	"github.com/tribehq/platform/models"
	"go.mongodb.org/mongo-driver/bson"
)

//Restaurants returns a list of restaurants.
func (r *queryResolver) Restaurants(ctx context.Context, name *string, after *string, before *string, first *int, last *int) (*models.RestaurantConnection, error) {
	var items []*models.Restaurant
	var edges []*models.RestaurantEdge
	filter := bson.D{}
	limit := 25
	items, totalCount, hasPrevious, hasNext, err := models.GetRestaurants(filter, limit, after, before, first, last)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		edge := &models.RestaurantEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(item.ID.Hex())),
			Node:   item,
		}
		edges = append(edges, edge)
	}

	pageInfo := getPageInfo(edges[0].Cursor, edges[len(edges)-1].Cursor, len(edges), hasNext, hasPrevious)

	itemList := &models.RestaurantConnection{TotalCount: int(totalCount), Edges: edges, Nodes: items, PageInfo: pageInfo}
	return itemList, nil
}

type restaurantResolver struct{ *Resolver }
