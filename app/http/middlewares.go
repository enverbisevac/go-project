package http

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/enverbisevac/go-project/app"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err != nil {
				s.error(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (s *Server) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")

		token := r.Header.Get("Authorization")
		if token == "" {
			next.ServeHTTP(w, r)
			return
		}

		claims, err := s.jwt.Verify(r.Context(), token)
		if err != nil {
			s.error(w, r, err)
			return
		}

		r = contextSetAuthUser(r, &claims.AuthUser)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) requireAuthUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authenticatedUser := contextGetAuthUser(r)

		if authenticatedUser == nil {
			s.authRequired(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) authorize(next http.Handler, permission string, resourceParamName string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := contextGetAuthUser(r)

		if session == nil {
			s.authzRequired(w, r)
			return
		}

		params := httprouter.ParamsFromContext(r.Context())
		resID := params.ByName(resourceParamName)

		if ok, err := s.authorizer.Authorize(r.Context(), session, app.PermissionCheck{
			Permission: permission,
			ResourceID: &resID,
		}); err != nil || !ok {
			s.authzRequired(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) requireBasicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, plaintextPassword, ok := r.BasicAuth()
		if !ok {
			s.basicAuthRequired(w, r)
			return
		}

		_, err := s.authenticator.Authenticate(r.Context(), app.Credentials{
			Email:    app.Email(username),
			Password: app.Password(plaintextPassword),
		})

		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			s.basicAuthRequired(w, r)
			return
		case err != nil:
			s.error(w, r, err)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) loggerHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := log.With().Logger()
		r = r.WithContext(l.WithContext(r.Context()))
		next.ServeHTTP(w, r)
	})
}
