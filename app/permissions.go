package app

import (
	"context"
)

const (
	//
	// Users
	//
	PermissionCreateUser string = "create_user"
	PermissionViewUser   string = "view_user"
	PermissionUpdateUser string = "update_user"
	PermissionDeleteUser string = "delete_user"
	//
	// Roles
	//
	PermissionCreateRole string = "create_role"
	PermissionViewRole   string = "view_role"
	PermissionUpdateRole string = "update_role"
	PermissionDeleteRole string = "delete_role"
)

type PermissionCheck struct {
	Permission string  `json:"permission"`
	ResourceID *string `json:"resource_id"`
}

type Authorizer interface {
	Authorize(ctx context.Context, session Session, permission ...PermissionCheck) (bool, error)
}

var Permissions = []struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Deprecated bool   `json:"deprecated"`
}{
	// Users
	{ID: PermissionCreateUser, Name: "Create a user"},
	{ID: PermissionViewUser, Name: "Get a user data"},
	{ID: PermissionUpdateUser, Name: "Update a user data"},
	{ID: PermissionDeleteUser, Name: "Delete a user"},
	// Roles
	{ID: PermissionCreateRole, Name: "Create a role"},
	{ID: PermissionViewRole, Name: "Get role"},
	{ID: PermissionUpdateRole, Name: "Update a role"},
	{ID: PermissionDeleteRole, Name: "Delete a role"},
}
