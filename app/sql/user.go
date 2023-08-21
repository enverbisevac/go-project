package sql

import (
	"context"
	"strings"
	"time"

	"github.com/dchest/uniuri"
	"github.com/enverbisevac/go-project/app"
)

const (
	selectUsers = `
	SELECT
		user_id,
		user_active,
		user_created,
		user_modified,
		user_email,
		user_full_name,
		user_is_admin,
		user_date_joined,
		user_last_login,
		user_salt
	FROM users
	`
)

func (ds *DataSource) InsertUser(ctx context.Context, in *app.User) error {
	if err := in.Validate(); err != nil {
		return err
	}

	const query = `
	INSERT INTO users(
		user_id,
		user_active,
		user_created,
		user_email,
		user_full_name,
		user_is_admin,
		user_date_joined,
		user_salt,
		user_hashed_password
	) VALUES (
		:user_id,
		:user_active,
		:user_created,
		:user_email,
		:user_full_name,
		:user_is_admin,
		:user_date_joined,
		:user_salt,
		:user_hashed_password
	)
	`

	in.Created = time.Now().Unix()
	in.DateJoined = in.Created

	if in.Password != "" {
		hashedPassword, err := passwordHash(in.Password)
		if err != nil {
			return app.ErrInternal("hash password error", err)
		}

		in.Password = app.Password(hashedPassword)
	}

	in.Salt = uniuri.NewLen(uniuri.UUIDLen)

	return insertSQL(ctx, ds, query, in)
}

func (ds *DataSource) GetUser(ctx context.Context, filter app.UserFilter) (*app.User, error) {
	query := selectUsers + `
	WHERE user_id = $1
		OR LOWER(user_email) = LOWER($2)
	LIMIT 1
	`

	user := &app.User{}
	if err := ds.GetContext(ctx,
		user,
		query,
		filter.ID,
		filter.Email,
	); err != nil {
		return nil, wrapError(err, "user", filter.String())
	}
	return user, nil
}

type userCredentials struct {
	ID       string  `db:"user_id"`
	Active   bool    `db:"user_active"`
	Password *string `db:"user_hashed_password"`
	Salt     string  `db:"user_salt"`
}

// never expose this method
func (ds *DataSource) getUserCredentials(ctx context.Context, filter app.UserFilter) (*userCredentials, error) {
	const query = `
	SELECT
		user_id,
		user_active,
		user_hashed_password,
		user_salt
	FROM users
	WHERE user_id = $1
		OR LOWER(user_email) = LOWER($2)
	LIMIT 1
	`
	userCreds := &userCredentials{}
	err := ds.GetContext(ctx, userCreds, query, filter.ID, filter.Email)
	if err != nil {
		return nil, err
	}
	return userCreds, nil
}

func (ds *DataSource) FindUsers(ctx context.Context) ([]app.User, error) {
	var users []app.User

	err := ds.SelectContext(ctx, &users, selectUsers)
	if err != nil {
		return []app.User{}, app.ErrInternal("failed to retrieve users", err)
	}

	return users, nil
}

func (ds *DataSource) FindAdmins(ctx context.Context) ([]app.User, error) {
	const query = selectUsers + `
	WHERE user_is_admin = 1
	`
	rows := make([]app.User, 0, 20)
	if err := ds.SelectContext(ctx, &rows, query); err != nil {
		return nil, app.ErrInternal("failed to get admin users", err)
	}
	return rows, nil
}

func (ds *DataSource) UpdateUserPassword(ctx context.Context, filter app.UserFilter, password app.Password) error {
	const query = `
	UPDATE users
	SET
		user_hashed_password = $1
	WHERE user_id = $2 OR LOWER(user_email) = LOWER($3)
	`

	hashedPassword, err := passwordHash(password)
	if err != nil {
		return err
	}

	return updateSQL(ctx, ds, query, hashedPassword, filter.ID, filter.Email)
}

func (ds *DataSource) UpdateUser(ctx context.Context, user *app.User, id string) error {
	const query = `
	UPDATE users
	SET
		user_active = :user_active,
		user_modified = :user_modified,
		user_email = :user_email,
		user_full_name = :user_full_name,
		user_is_admin = :user_is_admin
	WHERE user_id = $1
	`

	return updateSQL(ctx, ds, query, user, id)
}

func (ds *DataSource) DeleteUser(ctx context.Context, filter app.UserFilter) error {
	const query = `
	DELETE FROM users
	WHERE user_id = $1 OR LOWER(user_email) = LOWER($2)
	`

	return deleteSQL(ctx, ds, query, filter.ID, filter.Email)
}

func (ds *DataSource) updateUserLastLogin(ctx context.Context, filter app.UserFilter) error {
	const query = `
	UPDATE users
	SET
		user_last_login = $1
	WHERE user_id = $2 OR LOWER(user_email) = LOWER($3)
	`
	return updateSQL(ctx, ds, query, time.Now().Unix(), filter.ID, filter.Email)
}

func (ds *DataSource) InsertUserRole(ctx context.Context, in *app.UserRole) error {
	const query = `
	INSERT INTO user_roles(
		user_role_user_id,
		user_role_role_id,
		user_role_created
	) VALUES (
		:user_role_user_id,
		:user_role_role_id,
		:user_role_created
	)
	`
	return insertSQL(ctx, ds, query, in)
}

func (ds *DataSource) DeleteUserRoles(ctx context.Context, userID string) error {
	const query = `
	DELETE FROM user_roles
	WHERE user_role_user_id = $1
	`
	return deleteSQL(ctx, ds, query, userID)
}

func (ds *DataSource) GetUserRoles(ctx context.Context, userID string) ([]app.UserRole, error) {
	const query = `
		SELECT 
		user_role_user_id,
		user_role_role_id,
		user_role_created
		FROM user_roles 
		WHERE user_role_user_id = $1`

	rows := make([]app.UserRole, 0, 20)
	if err := ds.SelectContext(ctx, &rows, query, userID); err != nil {
		return nil, app.ErrInternal("failed to get user roles with user id in %v", userID, err)
	}
	return rows, nil
}

// User service methods

func (db *DB) AddUser(ctx context.Context, user *app.UserAggregate) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.InsertUser(ctx, &user.User)
	if err != nil {
		return err
	}

	// add user roles
	for _, roleID := range user.Roles {
		err = tx.InsertUserRole(ctx, &app.UserRole{
			UserID: user.ID,
			RoleID: roleID,
		})
		if err != nil {
			switch {
			case strings.Contains(err.Error(), "FOREIGN KEY constraint failed"):
				return app.ErrInvalid("role %s not found", roleID, err)
			}
			return err
		}
	}

	// add user permissions
	for _, permission := range user.Permissions {
		err = tx.InsertPermission(ctx, &app.Permission{
			UserID:       &user.ID,
			PermissionID: permission.Permission,
			ResourceID:   permission.ResourceID,
		})
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (db *DB) UpdateUser(ctx context.Context, user *app.UserAggregate) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.UpdateUser(ctx, &user.User, user.ID)
	if err != nil {
		return err
	}

	// delete user roles
	err = tx.DeleteUserRoles(ctx, user.ID)
	if err != nil && app.ErrorStatus(err) != app.StatusNotFound {
		return err
	}

	// add user roles
	for _, roleID := range user.Roles {
		err = tx.InsertUserRole(ctx, &app.UserRole{
			UserID: user.ID,
			RoleID: roleID,
		})
		if err != nil {
			return err
		}
	}

	// delete permissions
	err = tx.DeletePermissions(ctx, app.PermissionFilter{
		UserID: user.ID,
	})
	if err != nil {
		return err
	}

	// add user permissions
	for _, permission := range user.Permissions {
		err = tx.InsertPermission(ctx, &app.Permission{
			UserID:       &user.ID,
			PermissionID: permission.Permission,
			ResourceID:   permission.ResourceID,
		})
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (db *DB) GetUser(ctx context.Context, filter app.UserFilter) (app.UserAggregate, error) {
	tx, err := db.Beginx()
	if err != nil {
		return app.UserAggregate{}, err
	}
	defer tx.Rollback()

	user, err := tx.GetUser(ctx, filter)
	if err != nil {
		return app.UserAggregate{}, err
	}

	permissions, err := tx.GetPermissions(ctx, app.PermissionFilter{
		UserID: user.ID,
	})
	if err != nil {
		return app.UserAggregate{}, err
	}

	perms := make([]app.PermissionCheck, len(permissions))
	for i, perm := range permissions {
		perms[i] = app.PermissionCheck{
			Permission: perm.PermissionID,
			ResourceID: perm.ResourceID,
		}
	}

	roles, err := tx.GetUserRoles(ctx, user.ID)
	if err != nil {
		return app.UserAggregate{}, err
	}

	rolesIDs := make([]string, len(roles))
	for i, role := range roles {
		rolesIDs[i] = role.RoleID
	}

	return app.UserAggregate{
		User:        *user,
		Permissions: perms,
		Roles:       rolesIDs,
	}, nil
}

func (db *DB) DeleteUser(ctx context.Context, filter app.UserFilter) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	user, err := tx.GetUser(ctx, filter)
	if err != nil {
		return err
	}

	err = tx.DeletePermissions(ctx, app.PermissionFilter{
		UserID: user.ID,
	})
	if err != nil {
		return err
	}

	err = tx.DeleteUser(ctx, filter)
	if err != nil {
		return err
	}

	return tx.Commit()
}
