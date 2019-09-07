/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package resolvers

import (
	"context"
	"errors"
	"github.com/jinzhu/copier"
	log "github.com/sirupsen/logrus"
	"github.com/tribehq/platform/lib/audit_log"
	"github.com/tribehq/platform/models"
	"github.com/tribehq/platform/utils"
	"github.com/tribehq/platform/utils/auth"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//AppLatestVersion gives the latest version of the app according to its ID
func (r *queryResolver) AppLatestVersion(ctx context.Context, oSType string, packageID primitive.ObjectID, currentVersion string) (*models.AppVersion, error) {
	appVersion, err := models.GetAppLatestVersion()
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	return appVersion, nil
}

//AddAppVersion adds a new app version
func (r *mutationResolver) AddAppVersion(ctx context.Context, input models.AddAppVersionInput) (*models.AppVersion, error) {
	appVersion := &models.AppVersion{}
	_ = copier.Copy(&appVersion, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	appVersion.CreatedBy = user.ID
	appVersion, err = models.CreateAppVersion(*appVersion)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Created, user.ID.Hex(), appVersion.ID.Hex(), "app version", appVersion, nil, ctx)
	return appVersion, nil
}

//UpdateAppVersion updates an existing app version
func (r *mutationResolver) UpdateAppVersion(ctx context.Context, input models.UpdateAppVersionInput) (*models.AppVersion, error) {
	appVersion := &models.AppVersion{}
	appVersion = models.GetAppVersionByID(input.ID.Hex())
	_ = copier.Copy(&appVersion, &input)
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	appVersion.CreatedBy = user.ID
	appVersion, err = models.UpdateAppVersion(appVersion)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Updated, user.ID.Hex(), appVersion.ID.Hex(), "app version", appVersion, nil, ctx)
	return appVersion, nil
}

//DeleteAppVersion deleted an existing app version
func (r *mutationResolver) DeleteAppVersion(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	res, err := models.DeleteAppVersionByID(id.Hex())
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deleted, user.ID.Hex(), id.Hex(), "app version", nil, nil, ctx)
	return &res, err
}

//DeactivateAppVersion deactivates an app version
func (r *mutationResolver) DeactivateAppVersion(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	appVersion := models.GetAppVersionByID(id.Hex())
	if appVersion.ID.IsZero() {
		return utils.PointerBool(false), errors.New("app version not found")
	}
	appVersion.IsActive = false
	_, err := models.UpdateAppVersion(appVersion)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Deactivated, user.ID.Hex(), id.Hex(), "app version", nil, nil, ctx)
	return utils.PointerBool(false), nil
}

//ActivateAppVersion activates an app version
func (r *mutationResolver) ActivateAppVersion(ctx context.Context, id primitive.ObjectID) (*bool, error) {
	appVersion := models.GetAppVersionByID(id.Hex())
	if appVersion.ID.IsZero() {
		return utils.PointerBool(false), errors.New("app version not found")
	}
	appVersion.IsActive = true
	_, err := models.UpdateAppVersion(appVersion)
	if err != nil {
		return utils.PointerBool(false), err
	}
	user, err := auth.ForContext(ctx)
	if err != nil {
		return nil, err
	}
	//Update audit log
	go audit_log.NewAuditLogWithCtx(models.Activated, user.ID.Hex(), id.Hex(), "app version", nil, nil, ctx)
	return utils.PointerBool(true), nil
}
