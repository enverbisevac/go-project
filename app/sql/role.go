package sql

import (
	"context"

	"github.com/enverbisevac/go-project/app"
	"github.com/jmoiron/sqlx"
)

const (
	selectRoles = `--sql
	SELECT	
		role_id,
		role_created,
		role_modified,
		role_name
	FROM roles
	`
)

func (ds *DataSource) InsertRole(ctx context.Context, in *app.Role) error {
	const query = `
	INSERT INTO roles(
		role_id,
		role_created,
		role_name
	) VALUES (
		:role_id,
		:role_created,
		:role_name
	)
	`
	if err := in.Validate(); err != nil {
		return err
	}
	return insertSQL(ctx, ds, query, in)
}

func (ds *DataSource) getRole(ctx context.Context, filter *app.IDOrNameFilter) (*app.Role, error) {
	const query = selectRoles + `
	WHERE role_id = $1 OR LOWER(role_name) = LOWER($2)
	`
	role, err := getSQL[app.Role](ctx,
		ds,
		query,
		filter.ID,
		filter.Name,
	)
	if err != nil {
		return nil, wrapError(err, "role", filter.String())
	}
	return role, nil
}

func (ds *DataSource) FindRoles(ctx context.Context, ids ...string) ([]app.Role, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	query := selectRoles

	if len(ids) > 0 {
		query += " WHERE role_id in ($1) "
	}

	query, args, err := sqlx.In(query, ToSliceOfAny(ids)...)
	if err != nil {
		return nil, err
	}

	var roles []app.Role
	if err := ds.SelectContext(ctx, &roles, query, args...); err != nil {
		return []app.Role{}, err
	}

	return roles, nil
}

func (ds *DataSource) UpdateRole(ctx context.Context, role *app.Role, filter app.IDOrNameFilter) error {
	if err := role.Validate(); err != nil {
		return err
	}

	const query = `--sql
	UPDATE roles
	SET
		role_id,
		role_created,
		role_modified,
		role_name
	WHERE role_id = $1 OR LOWER(role_name) = LOWER($2)
	`

	return updateSQL(ctx, ds, query, filter.ID, filter.Name)
}

func (ds *DataSource) DeleteRole(ctx context.Context, filter *app.IDOrNameFilter) error {
	const query = `--sql
	DELETE FROM roles
	WHERE role_id = $1 OR LOWER(role_name) = LOWER($2)
	`
	return deleteSQL(ctx, ds, query, filter.ID, filter.Name)
}

func (db *DB) AddRole(ctx context.Context, in *app.RoleAggregate) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := tx.InsertRole(ctx, &in.Role); err != nil {
		return err
	}

	// add user permissions
	for _, permission := range in.Permissions {
		err = tx.InsertPermission(ctx, &app.Permission{
			RoleID:       &in.ID,
			PermissionID: permission.Permission,
			ResourceID:   permission.ResourceID,
		})
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (db *DB) UpdateRole(ctx context.Context, in *app.RoleAggregate, filter app.IDOrNameFilter) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.UpdateRole(ctx, &in.Role, filter)
	if err != nil {
		return err
	}

	// delete user permissions
	err = tx.DeletePermissions(ctx, app.PermissionFilter{
		RoleID: in.ID,
	})
	if err != nil {
		return err
	}

	// add user permissions
	for _, permission := range in.Permissions {
		err = tx.InsertPermission(ctx, &app.Permission{
			RoleID:       &in.ID,
			PermissionID: permission.Permission,
			ResourceID:   permission.ResourceID,
		})
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (db *DB) GetRole(ctx context.Context, filter *app.IDOrNameFilter) (app.RoleAggregate, error) {
	tx, err := db.BeginReadable()
	if err != nil {
		return app.RoleAggregate{}, err
	}
	defer tx.Rollback()

	role, err := tx.getRole(ctx, filter)
	if err != nil {
		return app.RoleAggregate{}, err
	}

	permissions, err := tx.GetPermissions(ctx, app.PermissionFilter{
		RoleID: role.ID,
	})
	if err != nil && app.ErrorStatus(err) != app.StatusNotFound {
		return app.RoleAggregate{}, err
	}

	permsChecks := make([]app.PermissionCheck, len(permissions))
	for i, perm := range permissions {
		permsChecks[i] = app.PermissionCheck{
			Permission: perm.PermissionID,
			ResourceID: perm.ResourceID,
		}
	}

	return app.RoleAggregate{
		Role:        *role,
		Permissions: permsChecks,
	}, nil
}

func (db *DB) DeleteRole(ctx context.Context, filter *app.IDOrNameFilter) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	role, err := tx.getRole(ctx, filter)
	if err != nil {
		return err
	}

	err = tx.DeletePermissions(ctx, app.PermissionFilter{
		RoleID: role.ID,
	})
	if err != nil {
		return err
	}

	err = tx.DeleteRole(ctx, filter)
	if err != nil {
		return err
	}

	return tx.Commit()
}
