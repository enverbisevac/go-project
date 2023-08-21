package sql

import (
	"context"
	"reflect"

	"github.com/enverbisevac/go-project/app"
	"github.com/jmoiron/sqlx"
)

func (ds *DataSource) InsertPermission(ctx context.Context, in *app.Permission) error {
	if err := in.Validate(); err != nil {
		return err
	}

	const query = `--sql
	INSERT INTO permissions(
		permission_user_id,
		permission_role_id,
		permission_id,
		permission_resource_id,
		permission_created
	) VALUES (
		:permission_user_id,
		:permission_role_id,
		:permission_id,
		:permission_resource_id,
		:permission_created
	)
	`

	return insertSQL(ctx, ds, query, in)
}

func (ds *DataSource) DeletePermissions(ctx context.Context, filter app.PermissionFilter) error {
	const query = `--sql
	DELETE FROM permissions
	WHERE permission_user_id = $1 OR permission_role_id = $2
	`
	return deleteSQL(ctx, ds, query, filter.UserID, filter.RoleID)
}

func (ds *DataSource) GetPermissions(
	ctx context.Context,
	filters ...app.PermissionFilter,
) ([]app.Permission, error) {
	// if no filters prvided dont do anything
	if len(filters) == 0 {
		return []app.Permission{}, nil
	}

	var (
		permissions []app.Permission
		field       string
		value       []any
	)

	userID := make([]any, 0, 1)
	roleIDs := make([]any, 0, 16)
	for _, filter := range filters {
		switch {
		case filter.UserID != "":
			if len(userID) == 0 {
				userID = append(userID, filter.UserID)
			}
		case filter.RoleID != "":
			roleIDs = append(roleIDs, filter.RoleID)
		}
	}

	field = "permission_user_id"
	value = userID
	if len(roleIDs) > 0 {
		field = "permission_role_id"
		value = roleIDs
	}

	query := `--sql
	SELECT
		permission_user_id,
		permission_role_id,
		permission_id,
		permission_resource_id,
		permission_created
	FROM permissions
	WHERE ` + field + ` in ($1)
	`

	query, args, err := sqlx.In(query, value...)
	if err != nil {
		return nil, err
	}

	if err := ds.SelectContext(ctx, &permissions, query, args...); err != nil {
		return nil, app.ErrInternal("failed to get user permissions with %s = %s",
			field, reflect.TypeOf(value).Elem(), err)
	}
	return permissions, nil
}

func (db *DB) CheckPermissions(ctx context.Context, userID string, permissions ...app.PermissionCheck) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	tx, err := db.Beginx()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	user, err := tx.GetUser(ctx, app.UserFilter{
		ID: userID,
	})
	if err != nil {
		return false, err
	}

	if user.IsAdmin {
		return true, nil
	}

	permChecks, err := tx.GetPermissions(ctx, app.PermissionFilter{
		UserID: user.ID,
	})
	if err != nil {
		return false, err
	}

	userRoles, err := tx.GetUserRoles(ctx, user.ID)
	if err != nil {
		return false, err
	}

	rolesIDs := make([]string, len(userRoles))
	for i, role := range userRoles {
		rolesIDs[i] = role.RoleID
	}

	roles, err := tx.FindRoles(ctx, rolesIDs...)
	if err != nil {
		return false, err
	}

	args := make([]app.PermissionFilter, len(roles))
	for i, role := range roles {
		args[i] = app.PermissionFilter{
			RoleID: role.ID,
		}
	}

	perms, err := tx.GetPermissions(ctx, args...)
	if err != nil {
		return false, err
	}

	permChecks = append(permChecks, perms...)

	for _, permission := range permissions {
		for _, pc := range permChecks {
			if pc.PermissionID != permission.Permission {
				continue
			}
			if permission.ResourceID != nil && *permission.ResourceID != "" &&
				pc.ResourceID == nil || *pc.ResourceID != *permission.ResourceID {
				continue
			}
			return true, nil
		}
	}
	return false, nil
}
