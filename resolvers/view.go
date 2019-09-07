/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package resolvers

import (
	"github.com/tribehq/platform/models"
	"golang.org/x/net/context"
)

//TODO
//GodsView returns godsview
func (r *queryResolver) GodsView(ctx context.Context, vehicleStatusType *models.VehicleStatusType, latitude *float64, longitude *float64, after *string, before *string, first *int, last *int) (*models.GodsViewConnection, error) {
	panic("implement me")
}

//TODO
//HeatView returns heatview
func (r *queryResolver) HeatView(ctx context.Context, latitude *float64, longitude *float64, after *string, before *string, first *int, last *int) (*models.HeatViewConnection, error) {
	panic("implement me")
}
