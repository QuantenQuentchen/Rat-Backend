package api

import (
	"RatBackend/db"
	"RatBackend/models"
)

func GetUserPermissions(userID string) (models.PermFlag, error) {
	roles, err := db.GetUserRoles(userID)
	if err != nil {
		return 0, err
	}
	var permissions uint64
	for _, role := range roles {
		permissions = permissions | role.Permissions
	}
	return models.PermFlag(permissions), nil
}

func GetUserRoles(userID string) ([]models.Role, error) {
	roles, err := db.GetUserRoles(userID)
	if err != nil {
		return nil, err
	}
	return roles, nil
}
