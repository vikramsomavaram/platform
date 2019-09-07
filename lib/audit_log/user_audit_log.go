/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package audit_log

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/tribehq/platform/models"
	"time"
)

//NewAuditLog inserts a new audit log into database
func NewAuditLog(actionType models.ActionType, userID string, objectID string, objectName string, objectData interface{}, actionMetadata map[string]string) {
	newAuditLog := &models.UserAuditLog{
		CreatedAt:      time.Now(),
		UserID:         userID,
		ObjectID:       objectID,
		ObjectName:     objectName,
		ObjectData:     objectData,
		ActionMetadata: actionMetadata,
		ActionType:     actionType,
	}
	_, err := models.CreateUserAuditLog(newAuditLog)
	if err != nil {
		log.Errorln(err)
	}
}

func NewAuditLogWithCtx(actionType models.ActionType, userID string, objectID string, objectName string, objectData interface{}, actionMetadata map[string]string, ctx context.Context) {
	//TODO: Get IP ADDR and access token details into ActionMetadata
	newAuditLog := &models.UserAuditLog{
		CreatedAt:      time.Now(),
		UserID:         userID,
		ObjectID:       objectID,
		ObjectName:     objectName,
		ObjectData:     objectData,
		ActionMetadata: actionMetadata,
		ActionType:     actionType,
	}
	_, err := models.CreateUserAuditLog(newAuditLog)
	if err != nil {
		log.Errorln(err)
	}
}
