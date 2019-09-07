/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package oauth

// IsRoleAllowed returns true if the role is allowed to use this service
func IsRoleAllowed(role string, allowedRoles []string) bool {
	for _, allowedRole := range allowedRoles {
		if role == allowedRole {
			return true
		}
	}
	return false
}
