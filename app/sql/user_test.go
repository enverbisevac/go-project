
package sql

import (
	"context"
	"testing"

	"github.com/enverbisevac/go-project/app"
)

func TestDataSource_InsertUser(t *testing.T) {
	db, teardown := setupTest(t)
	defer teardown()

	type fields struct {
		DAO DAO
	}
	type args struct {
		ctx context.Context
		in  *app.User
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		compare func(user *app.User) bool
	}{
		{
			name: "create user without ID",
			fields: fields{
				DAO: db,
			},
			args: args{
				ctx: context.Background(),
				in: &app.User{
					Active:   true,
					FullName: "User",
					Email:    "admin@domain.com",
					Password: "some password",
				},
			},
			wantErr: false,
			compare: func(user *app.User) bool {
				return user.ID != "" &&
					user.Active && user.FullName == "User"
			},
		},
		{
			name: "create user with ID",
			fields: fields{
				DAO: db,
			},
			args: args{
				ctx: context.Background(),
				in: &app.User{
					ID:       "qw420934ir0w9ei",
					Active:   true,
					FullName: "User 1",
					Email:    "admin@domain.com",
					Password: "some password",
				},
			},
			wantErr: false,
			compare: func(user *app.User) bool {
				if user.ID != "qw420934ir0w9ei" {
					t.Errorf("insert user got ID %s, expected ID %s", user.ID, "qw420934ir0w9ei")
					return false
				}
				return true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := &DataSource{
				DAO: tt.fields.DAO,
			}
			if err := ds.InsertUser(tt.args.ctx, tt.args.in); (err != nil) != tt.wantErr {
				t.Errorf("DataSource.InsertUser() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.compare(tt.args.in) {
				t.Errorf("DataSource.InsertUser() %s failed", tt.name)
			}
		})
	}
}

func TestDB_UpdateUser(t *testing.T) {
	db, teardown := setupTest(t)
	defer teardown()

	err := db.AddUser(context.Background(), &app.UserAggregate{
		User: app.User{
			ID:       "qw420934ir0w9ei",
			FullName: "User",
			Active:   true,
		},
		Permissions: []app.PermissionCheck{
			{Permission: app.PermissionCreateUser},
		},
	})
	if err != nil {
		t.Error(err)
	}

	type fields struct {
		DB *DB
	}
	type args struct {
		ctx  context.Context
		user *app.UserAggregate
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		compare func(user *app.User) bool
	}{
		{
			name: "happy path",
			fields: fields{
				DB: db,
			},
			args: args{
				ctx: context.Background(),
				user: &app.UserAggregate{
					User: app.User{
						ID:       "qw420934ir0w9ei",
						Active:   true,
						FullName: "User 1",
					},
					Permissions: []app.PermissionCheck{
						{Permission: app.PermissionViewUser},
					},
				},
			},
			wantErr: false,
			compare: func(user *app.User) bool {
				if user.FullName != "User 1" {
					t.Errorf("update user failed, got full name %s, expected %s", user.FullName, "User 1")
					return false
				}
				return true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := tt.fields.DB
			if err := db.UpdateUser(tt.args.ctx, tt.args.user); (err != nil) != tt.wantErr {
				t.Errorf("DB.UpdateUser() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.compare(&tt.args.user.User) {
				t.Errorf("DataSource.UpdateUser() %s failed", tt.name)
			}
		})
	}
}

func TestDataSource_DeleteUser(t *testing.T) {
	db, teardown := setupTest(t)
	defer teardown()

	err := db.AddUser(context.Background(), &app.UserAggregate{
		User: app.User{
			ID:       "qw420934ir0w9ei",
			FullName: "User",
			Active:   true,
		},
		Permissions: []app.PermissionCheck{
			{Permission: app.PermissionCreateUser},
		},
	})
	if err != nil {
		t.Error(err)
	}

	err = db.AddUser(context.Background(), &app.UserAggregate{
		User: app.User{
			ID:       "qw420934ir09080",
			FullName: "User 1",
			Active:   true,
		},
		Permissions: []app.PermissionCheck{
			{Permission: app.PermissionCreateUser},
		},
	})
	if err != nil {
		t.Error(err)
	}

	type fields struct {
		DAO DAO
	}
	type args struct {
		ctx    context.Context
		filter app.UserFilter
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		check   func() bool
	}{
		{
			name: "delete user by ID",
			fields: fields{
				DAO: db,
			},
			args: args{
				ctx: context.Background(),
				filter: app.UserFilter{
					ID: "qw420934ir09080",
				},
			},
			wantErr: false,
			check: func() bool {
				_, err := db.DataSource.GetUser(context.Background(), app.UserFilter{
					ID: "qw420934ir0w9ei",
				})
				if err != nil && app.ErrorStatus(err) == app.StatusNotFound {
					t.Errorf("qw420934ir0w9ei not found")
					return false
				}
				return true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := &DataSource{
				DAO: tt.fields.DAO,
			}
			if err := ds.DeleteUser(tt.args.ctx, tt.args.filter); (err != nil) != tt.wantErr {
				t.Errorf("DataSource.DeleteUser() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.check() {
				t.Errorf("DataSource.DeleteUser() %s failed", tt.name)
			}
		})
	}
}
