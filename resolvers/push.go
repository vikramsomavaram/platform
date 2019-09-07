/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package resolvers

import (
	"context"
	"github.com/tribehq/platform/models"
)

//TODO
//SendPushNotification returns a notification
func (r *mutationResolver) SendPushNotification(ctx context.Context, input models.PushNotificationInput) (*models.PushNotification, error) {
	//notification, err := models.GetPushNotificationByID(input.ID.Hex())
	//if err != nil {
	//	log.Errorln(err)
	//	return nil, err
	//}
	//return notification, nil
	panic("implement me")
}
