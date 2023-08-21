package http

import (
	"context"
	"net/http"

	"github.com/enverbisevac/go-project/app"
)

type contextKey string

const (
	authUserContextKey = contextKey("authUser")
)

func contextSetAuthUser(r *http.Request, user *app.AuthUser) *http.Request {
	ctx := context.WithValue(r.Context(), authUserContextKey, user)
	return r.WithContext(ctx)
}

func contextGetAuthUser(r *http.Request) *app.AuthUser {
	user, ok := r.Context().Value(authUserContextKey).(*app.AuthUser)
	if !ok {
		return nil
	}

	return user
}
