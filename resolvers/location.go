/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package resolvers

import (
	"context"
	"github.com/tribehq/platform/lib/audit_log"
	"github.com/tribehq/platform/models"
	"github.com/tribehq/platform/utils/auth"
)

//UpdateProviderLocation updates provider location
func (r *mutationResolver) UpdateProviderLocation(ctx context.Context, latitude float64, longitude float64) (bool, error) {
	location := &models.ServiceProviderLocation{}
	location.Location = models.Location{Type: "Point", Coordinates: []float64{latitude, longitude}}
	location.ServiceProviderID = ""
	location, err := models.CreateServiceProviderLocation(location)
	if err != nil {
		return false, err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return false, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), location.ID.Hex(), "provider location", location, nil, ctx)
	return !location.ID.IsZero(), nil
}

//UpdateUserLocation updates user location
func (r *mutationResolver) UpdateUserLocation(ctx context.Context, latitude float64, longitude float64) (bool, error) {
	userLocation := &models.UserLocation{}
	userLocation.Location = models.Location{Type: "Point", Coordinates: []float64{latitude, longitude}}
	userLocation.UserID = ""
	userLocation, err := models.CreateUserLocation(userLocation)
	if err != nil {
		return false, err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return false, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), userLocation.ID.Hex(), "user location", userLocation, nil, ctx)
	return !userLocation.ID.IsZero(), nil
}
