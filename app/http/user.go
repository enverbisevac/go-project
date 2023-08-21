package http

import (
	"net/http"
	"strings"

	"github.com/enverbisevac/go-project/app"
	"github.com/julienschmidt/httprouter"
	"github.com/swaggest/openapi-go/openapi3"
)

func (s *Server) createUserHandler() http.HandlerFunc {
	// define openapi operation
	opCreate := createSecureOperation("users", "createUser", "Create a new user")

	success := s.createAPIResponses(&opCreate, app.UserAggregate{})
	handleError(s.reflector.SetRequest(&opCreate, app.UserAggregate{}, routes.createUser.method))
	handleError(s.reflector.Spec.AddOperation(routes.createUser.method, routes.createUser.getOAPI(), opCreate))

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		in := &app.UserAggregate{}

		err := DecodeJSON(w, r, in)
		if err != nil {
			s.invalidBody(w, r, "JSON", err)
			return
		}

		if err = in.Validate(); err != nil {
			s.error(w, r, err)
			return
		}

		existingUser, err := s.store.GetUser(ctx, app.UserFilter{
			Email: (*string)(&in.Email),
		})
		if err != nil && app.ErrorStatus(err) != app.StatusNotFound {
			s.error(w, r, err)
			return
		}

		if existingUser.ID != "" {
			s.error(w, r, app.ErrConflict("email is already taken"))
			return
		}

		err = s.store.AddUser(ctx, in)
		if err != nil {
			s.error(w, r, err)
			return
		}

		JSON(w, success, in)
	}
}

func (s *Server) updateUserHandler() http.HandlerFunc {
	// define openapi operation
	opUpdate := createSecureOperation("users", "updateUser", "Update existing user")
	opUpdate.Parameters = []openapi3.ParameterOrRef{
		{Parameter: paramID},
	}

	success := s.updateAPIResponses(&opUpdate, new(app.UserAggregate))
	handleError(s.reflector.SetRequest(&opUpdate, new(app.UserAggregate), routes.updateUser.method))
	handleError(s.reflector.Spec.AddOperation(routes.updateUser.method, routes.updateUser.getOAPI(), opUpdate))

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		params := httprouter.ParamsFromContext(r.Context())
		id := params.ByName(paramID.Name)
		in := &app.UserAggregate{}

		err := DecodeJSON(w, r, in)
		if err != nil {
			s.invalidBody(w, r, "JSON", err)
			return
		}

		if err := in.Validate(); err != nil {
			s.error(w, r, err)
			return
		}

		in.ID = id

		err = s.store.UpdateUser(ctx, in)
		if err != nil {
			s.error(w, r, err)
			return
		}

		JSON(w, success, in)
	}
}

func (s *Server) updateUserPasswordHandler() http.HandlerFunc {
	// define openapi operation
	type ChangePasswordRequest struct {
		ID       string       `json:"id"`
		Password app.Password `json:"password"`
	}

	opUpdatePassword := createSecureOperation("users", "updateUserPassword", "Update existing user")
	opUpdatePassword.Parameters = []openapi3.ParameterOrRef{
		{Parameter: paramID},
	}

	success := s.updateAPIResponses(&opUpdatePassword, nil)
	handleError(s.reflector.SetRequest(&opUpdatePassword, ChangePasswordRequest{}, routes.updatePassword.method))
	handleError(s.reflector.Spec.AddOperation(routes.updatePassword.method, routes.updatePassword.getOAPI(), opUpdatePassword))

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		params := httprouter.ParamsFromContext(r.Context())
		id := params.ByName(paramID.Name)
		in := &ChangePasswordRequest{}

		err := DecodeJSON(w, r, in)
		if err != nil {
			s.error(w, r, err)
			return
		}

		err = s.store.UpdateUserPassword(ctx, app.UserFilter{
			ID: id,
		}, in.Password)
		if err != nil {
			s.error(w, r, err)
			return
		}

		w.WriteHeader(success)
	}
}

func (s *Server) deleteUserHandler() http.HandlerFunc {
	// define openapi operation
	opDelete := createSecureOperation("users", "deleteUser", "Delete a user")
	opDelete.Parameters = []openapi3.ParameterOrRef{
		{Parameter: paramID},
	}

	success := s.deleteAPIResponses(&opDelete)
	handleError(s.reflector.Spec.AddOperation(routes.deleteUser.method, routes.deleteUser.getOAPI(), opDelete))

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		params := httprouter.ParamsFromContext(r.Context())
		id := params.ByName(paramID.Name)

		err := s.store.DeleteUser(ctx, app.UserFilter{
			ID: id,
		})
		if err != nil {
			s.error(w, r, err)
			return
		}

		w.WriteHeader(success)
	}
}

func (s *Server) getUserHandler() http.HandlerFunc {
	// define openapi operation
	opRead := createSecureOperation("users", "getUser", "Get user data by ID")
	opRead.Parameters = []openapi3.ParameterOrRef{
		{Parameter: paramID},
	}

	success := s.getAPIResponses(&opRead, app.UserAggregate{})
	handleError(s.reflector.Spec.AddOperation(routes.getUser.method, routes.getUser.getOAPI(), opRead))

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		session := contextGetAuthUser(r)
		params := httprouter.ParamsFromContext(r.Context())
		id := params.ByName(paramID.Name)

		if strings.ToLower(id) == "me" {
			id = session.UserID()
		}

		user, err := s.store.GetUser(ctx, app.UserFilter{
			ID: id,
		})
		if err != nil {
			s.error(w, r, err)
			return
		}

		JSON(w, success, user)
	}
}
