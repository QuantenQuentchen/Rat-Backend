package api

import (
	"RatBackend/db"
	"RatBackend/models"
)

type FunctionParameters interface {
	Nothing()
}

type CreateUserRequest struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
}

func (f CreateUserRequest) Nothing() {
	// This method is used to satisfy the FunctionParameters interface
}

type AssignRoleRequest struct {
	UserID       string `json:"user_id"`
	RoleID       string `json:"role_id"`
	IsTransferal bool   `json:"is_transferal"`
}

func (f AssignRoleRequest) Nothing() {
}

type ModifyRoleRequest struct {
	RoleID string      `json:"role_id"`
	Role   models.Role `json:"role"`
}

func (f ModifyRoleRequest) Nothing() {}

type CreateRoleRequest struct {
	Role models.Role `json:"role"`
}

func (f CreateRoleRequest) Nothing() {}

func CreateUser(request CreateUserRequest) error {
	userID := request.UserID
	name := request.Name
	return db.AddUser(userID, name)
}

func AssignRole(request AssignRoleRequest) error {
	binding := models.RoleBinding{
		UserId:     request.UserID,
		RoleId:     request.RoleID,
		Transferal: request.IsTransferal,
	}
	return db.AssignRoles(binding)
}

func CreateRole(request CreateRoleRequest) (string, error) {
	err := db.ValidateRole(request.Role)
	if err != nil {
		return "", err
	}
	id, err := db.AddRole(request.Role)
	if err != nil {
		return "", err
	}
	return string(id), nil
}

func ModifyRole(request ModifyRoleRequest) error {
	err := db.ValidateRole(request.Role)
	if err != nil {
		return err
	}
	return db.UpdateRole(request.RoleID, request.Role)
}
