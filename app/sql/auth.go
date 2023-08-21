package sql

import (
	"context"
	"errors"

	"github.com/enverbisevac/go-project/app"
	"github.com/enverbisevac/go-project/pkg/ptr"
	"golang.org/x/crypto/bcrypt"
)

func (ds *DataSource) Authenticate(ctx context.Context, creds app.Credentials) (app.AuthUser, error) {
	if err := creds.Validate(); err != nil {
		return app.AuthUser{}, err
	}

	userCreds, err := ds.getUserCredentials(ctx, app.UserFilter{
		Email: ptr.From(creds.Email.String()),
	})
	if err != nil {
		if app.ErrorStatus(err) == app.StatusNotFound {
			return app.AuthUser{}, app.ErrUnauthenticated("user %s doesn't exists", creds.Email, err)
		}
		return app.AuthUser{}, err
	}

	if !userCreds.Active {
		return app.AuthUser{}, app.ErrUnauthenticated("user %s is deactivated", creds.Email)
	}

	passwordMatches, err := passwordMatches(*userCreds.Password, creds.Password.String())
	if err != nil {
		return app.AuthUser{}, err
	}
	if !passwordMatches {
		return app.AuthUser{}, app.ErrUnauthenticated("wrong password")
	}

	err = ds.updateUserLastLogin(ctx, app.UserFilter{
		ID: userCreds.ID,
	})
	if err != nil {
		return app.AuthUser{}, err
	}

	return app.AuthUser{
		ID:   userCreds.ID,
		Salt: userCreds.Salt,
	}, nil
}

func (db *DB) Authorize(ctx context.Context, session app.Session, permissions ...app.PermissionCheck) (bool, error) {
	return db.CheckPermissions(ctx, session.UserID(), permissions...)
}

func passwordHash[T ~string](plaintextPassword T) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return "", app.ErrInternal("hash password failed", err)
	}

	return string(hashedPassword), nil
}

func passwordMatches(hashedPassword, plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, app.ErrInternal("compare passwords failed", err)
		}
	}

	return true, nil
}
