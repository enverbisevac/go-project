package app

import (
	"context"
)

type Authenticator interface {
	Authenticate(ctx context.Context, credentials Credentials) (AuthUser, error)
}

type Storage interface {
	//
	// Users
	//
	AddUser(ctx context.Context, user *UserAggregate) error
	GetUser(ctx context.Context, filter UserFilter) (UserAggregate, error)
	UpdateUser(ctx context.Context, user *UserAggregate) error
	UpdateUserPassword(ctx context.Context, filter UserFilter, password Password) error
	DeleteUser(ctx context.Context, filter UserFilter) error
	FindUsers(ctx context.Context) ([]User, error)
	FindAdmins(ctx context.Context) ([]User, error)
	//
	// Roles
	//
	AddRole(ctx context.Context, in *RoleAggregate) error
	GetRole(ctx context.Context, filter *IDOrNameFilter) (RoleAggregate, error)
	UpdateRole(ctx context.Context, in *RoleAggregate, filter IDOrNameFilter) error
	DeleteRole(ctx context.Context, filter *IDOrNameFilter) error
	FindRoles(ctx context.Context, ids ...string) ([]Role, error)
}
