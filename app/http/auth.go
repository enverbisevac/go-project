package http

import (
	"net/http"
	"time"

	"github.com/enverbisevac/go-project/app"
)

func (s *Server) loginHandler() http.HandlerFunc {
	const success = http.StatusOK
	// define openapi operation
	opLogin := createOperation("users", "login", "authentication endpoint")

	handleError(s.reflector.SetRequest(&opLogin, new(app.Credentials), routes.login.method))
	handleError(s.reflector.SetJSONResponse(&opLogin, new(app.TokenData), success))
	handleError(s.reflector.SetJSONResponse(&opLogin, new(ErrorResponse), http.StatusBadRequest))
	handleError(s.reflector.SetJSONResponse(&opLogin, new(ErrorResponse), http.StatusUnauthorized))
	handleError(s.reflector.SetJSONResponse(&opLogin, new(ErrorResponse), http.StatusInternalServerError))
	handleError(s.reflector.Spec.AddOperation(routes.login.method, routes.login.path, opLogin))

	return func(w http.ResponseWriter, r *http.Request) {
		in := app.Credentials{}

		err := DecodeJSON(w, r, &in)
		if err != nil {
			s.invalidBody(w, r, "JSON", err)
			return
		}

		if err := in.Validate(); err != nil {
			s.error(w, r, err)
			return
		}

		authUser, err := s.authenticator.Authenticate(r.Context(), in)
		if err != nil {
			s.error(w, r, err)
			return
		}

		jwtBytes, expiry, err := s.jwt.Generate(&authUser)
		if err != nil {
			s.error(w, r, err)
			return
		}

		output := app.TokenData{
			Token:       string(jwtBytes),
			TokenExpire: expiry.Format(time.RFC3339),
		}

		err = JSON(w, success, output)
		if err != nil {
			s.error(w, r, err)
		}
	}
}
