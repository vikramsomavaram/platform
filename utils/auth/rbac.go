/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package auth

import (
	"github.com/mikespook/gorbac"
	log "github.com/sirupsen/logrus"
)

var (
	// RBAC return global gorbac instance.
	RBAC *gorbac.RBAC
)

// InitRBAC returns initializes the rbac instance.
func InitRBAC() *gorbac.RBAC {

	rbac := gorbac.New()

	roles := []*gorbac.StdRole{}
	//for _,role := range CoreRoles() {
	//	gorbac.NewStdRole(role)
	//}

	//default admin role
	//adminRole := gorbac.NewStdRole("admin")
	//
	//coreRolePermissions := CoreRolePermissions()
	//
	//for resourceGroup, resourcePermissions := range coreRolePermissions {
	//	for _, resourcePermission := range resourcePermissions {
	//		//create layered permission and add to the default admin role
	//		_ = adminRole.Assign(gorbac.NewLayerPermission(resourceGroup + ":" + resourcePermission))
	//	}
	//}

	//get all roles from database
	//dbRoles := models.GetAllUserRoles()
	//for _, dbRole := range dbRoles {
	//	role := gorbac.NewStdRole(dbRole.Name)
	//	for _, permission := range dbRole.Permissions {
	//		for _, action := range permission.Actions {
	//			err := role.Assign(gorbac.NewLayerPermission(permission.Resource + ":" + action))
	//			if err != nil {
	//				log.Errorln(err)
	//			}
	//		}
	//	}
	//	roles = append(roles, role)
	//}

	//Finally add all roles to rbac
	for _, role := range roles {
		err := rbac.Add(role)
		if err != nil {
			log.Errorln(err)
		}
	}

	RBAC = rbac
	return rbac
}
