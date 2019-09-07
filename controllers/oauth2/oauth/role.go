/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package oauth

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/tribehq/platform/models"
)

var (
	// ErrRoleNotFound ...
	ErrRoleNotFound = errors.New("role not found")
)

// FindRoleByID looks up a role by ID and returns it
func FindRoleByID(id string) (*models.UserRole, error) {
	role, err := models.GetUserRoleByID(id)
	if err != nil {
		log.Error(err)
	}
	if role.ID.IsZero() {
		return nil, ErrRoleNotFound
	}
	return role, nil
}
