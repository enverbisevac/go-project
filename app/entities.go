package app

import (
	"github.com/jaevor/go-nanoid"
)

type Credentials struct {
	Email    Email    `json:"email,writeOnly"`
	Password Password `json:"password,writeOnly"`
}

func (c *Credentials) Validate() error {
	if err := c.Email.Validate(); err != nil {
		return err
	}

	if err := c.Password.Validate(); err != nil {
		return err
	}

	return nil
}

type TokenData struct {
	Token       string `json:"token"`
	TokenExpire string `json:"token_expire"`
}

type User struct {
	ID         string   `db:"user_id" json:"id"`
	Active     bool     `db:"user_active" json:"active"`
	Created    int64    `db:"user_created" json:"created,readOnly"`
	Modified   *int64   `db:"user_modified" json:"modified,readOnly"`
	Email      Email    `db:"user_email" json:"email"`
	FullName   string   `db:"user_full_name" json:"full_name"`
	IsAdmin    bool     `db:"user_is_admin" json:"is_admin"`
	DateJoined int64    `db:"user_date_joined" json:"date_joined"`
	LastLogin  *int64   `db:"user_last_login" json:"last_login"`
	Salt       string   `db:"user_salt" json:"-"`
	Password   Password `db:"user_hashed_password" json:"password,writeOnly"`
}

func (u *User) Validate() error {
	if err := u.Email.Validate(); err != nil {
		return err
	}

	if err := u.Password.Validate(); err != nil {
		return err
	}

	if u.FullName == "" {
		return ErrFieldIsMandatory("full_name")
	}

	return nil
}

func (a *User) SetID(id any) {
	a.ID = id.(string)
}

func (a *User) GetID() any {
	return a.ID
}

func (a *User) SetCreated(val int64) {
	a.Created = val
}

func (a *User) SetModified(val int64) {
	a.Modified = &val
}

func (u *User) Generator() (func() any, error) {
	return generator()
}

type Role struct {
	ID       string `db:"role_id" json:"id"`
	Name     string `db:"role_name" json:"name"`
	Created  int64  `db:"role_created" json:"created,readOnly" readOnly:"true"`
	Modified *int64 `db:"role_modified" json:"modified,readOnly" readOnly:"true"`
}

func (r *Role) Validate() error {
	if r.Name == "" {
		return ErrFieldIsMandatory("name")
	}
	return nil
}

func (r *Role) SetID(id any) {
	r.ID = id.(string)
}

func (r *Role) GetID() any {
	return r.ID
}

func (r *Role) SetCreated(val int64) {
	r.Created = val
}

func (r *Role) SetModified(val int64) {
	r.Modified = &val
}

func (r *Role) Generator() (func() any, error) {
	return generator()
}

type UserRole struct {
	UserID  string `db:"user_role_user_id"`
	RoleID  string `db:"user_role_role_id"`
	Created int64  `db:"user_role_created"`
}

func (r *UserRole) Validate() error {
	if r.UserID == "" || r.RoleID == "" {
		return ErrFieldIsMandatory("user_id or role_id value is required")
	}

	return nil
}

func (r *UserRole) SetID(id any) {
	// dont to anything this is intersection table
}

func (r *UserRole) GetID() any {
	return r.UserID
}

func (r *UserRole) SetCreated(val int64) {
	r.Created = val
}

func (a *UserRole) SetModified(val int64) {
	// no implementation needed
}

func (r *UserRole) Generator() (func() any, error) {
	return nil, nil
}

type Permission struct {
	UserID       *string `db:"permission_user_id"`
	RoleID       *string `db:"permission_role_id"`
	PermissionID string  `db:"permission_id"`
	ResourceID   *string `db:"permission_resource_id"`
	Created      int64   `db:"permission_created"`
}

func (p *Permission) Validate() error {
	if p.UserID == nil && p.RoleID == nil {
		return ErrInvalid("please provide permission_user_id or permission_role_id")
	}

	if p.PermissionID == "" {
		return ErrInvalid("permission_id value is required")
	}

	return nil
}

func (p *Permission) SetID(id any) {
	// dont to anything this is intersection table
}

func (p *Permission) GetID() any {
	return p.UserID
}

func (p *Permission) SetCreated(val int64) {
	p.Created = val
}

func (p *Permission) SetModified(val int64) {
	// no implementation needed
}

func (p *Permission) Generator() (func() any, error) {
	return generator()
}

type UserPermission struct {
	UserID       string `db:"user_id"`
	PermissionID string `db:"user_permission_id"`
}

func generator() (func() any, error) {
	f, err := nanoid.Standard(21)
	return func() any {
		return f()
	}, err
}
