/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package resolvers

import (
	"context"
	"github.com/tribehq/platform/models"
)

type subscriptionResolver struct{ *Resolver }

//NearbyVehicles returns a list of nearby vehicles
func (r *subscriptionResolver) NearbyVehicles(ctx context.Context, latitude float64, longitude float64) (<-chan []*models.NearByVehicle, error) {
	panic("not implemented")
}

//SupportChatMessage gives a support chat message by its ID
func (r *subscriptionResolver) SupportChatMessage(ctx context.Context, chatID string) (<-chan *models.ChatMessage, error) {
	panic("not implemented")
}

//JobUpdates gives a list of job updates
func (r *subscriptionResolver) JobUpdates(ctx context.Context, jobID string) (<-chan *models.JobUpdate, error) {
	panic("not implemented")
}
