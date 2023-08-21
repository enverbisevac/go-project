package http

import (
	"net/http"

	"github.com/enverbisevac/go-project/app"
	"github.com/julienschmidt/httprouter"
	"github.com/swaggest/openapi-go/openapi3"
)

func (s *Server) createRoleHandler() http.HandlerFunc {
	// define openapi operation
	opCreate := createSecureOperation("roles", "createRole", "Create a new role")

	success := s.createAPIResponses(&opCreate, app.RoleAggregate{})
	handleError(s.reflector.SetRequest(&opCreate, app.RoleAggregate{}, routes.createRole.method))
	handleError(s.reflector.Spec.AddOperation(routes.createRole.method, routes.createRole.path, opCreate))

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		in := &app.RoleAggregate{}

		err := DecodeJSON(w, r, in)
		if err != nil {
			s.error(w, r, err)
			return
		}

		if err := in.Validate(); err != nil {
			s.error(w, r, err)
			return
		}

		existingRole, err := s.store.GetRole(ctx, &app.IDOrNameFilter{
			Name: in.Name,
		})
		if err != nil && app.ErrorStatus(err) != app.StatusNotFound {
			s.error(w, r, err)
			return
		}

		if existingRole.ID != "" {
			s.error(w, r, app.ErrConflict("role '%s' is already taken", in.Name))
			return
		}

		if err := s.store.AddRole(ctx, in); err != nil {
			s.error(w, r, err)
			return
		}

		JSON(w, success, in)
	}
}

func (s *Server) getRoleHandler() http.HandlerFunc {
	// define openapi operation
	opRead := createSecureOperation("roles", "getRole", "Get role data")
	opRead.Parameters = []openapi3.ParameterOrRef{
		{Parameter: paramID},
	}

	success := s.createAPIResponses(&opRead, app.RoleAggregate{})
	handleError(s.reflector.SetRequest(&opRead, app.RoleAggregate{}, routes.getRole.method))
	handleError(s.reflector.Spec.AddOperation(routes.getRole.method, routes.getRole.getOAPI(), opRead))

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		params := httprouter.ParamsFromContext(r.Context())
		id := params.ByName(paramID.Name)

		role, err := s.store.GetRole(ctx, &app.IDOrNameFilter{
			ID: id,
		})
		if err != nil {
			s.error(w, r, err)
			return
		}

		JSON(w, success, &role)
	}
}

func (s *Server) updateRoleHandler() http.HandlerFunc {
	// define openapi operation
	opUpdate := createSecureOperation("roles", "updateRole", "Update existing role")
	opUpdate.Parameters = []openapi3.ParameterOrRef{
		{Parameter: paramID},
	}

	success := s.updateAPIResponses(&opUpdate, app.RoleAggregate{})
	handleError(s.reflector.SetRequest(&opUpdate, app.RoleAggregate{}, routes.updateRole.method))
	handleError(s.reflector.Spec.AddOperation(routes.updateRole.method, routes.updateRole.getOAPI(), opUpdate))

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		params := httprouter.ParamsFromContext(r.Context())
		id := params.ByName(paramID.Name)
		in := &app.RoleAggregate{}

		err := DecodeJSON(w, r, in)
		if err != nil {
			s.error(w, r, err)
			return
		}

		in.ID = id

		err = s.store.UpdateRole(ctx, in, app.IDOrNameFilter{
			ID: id,
		})
		if err != nil {
			s.error(w, r, err)
			return
		}

		w.WriteHeader(success)
	}
}

func (s *Server) deleteRoleHandler() http.HandlerFunc {
	// define openapi operation
	opDelete := createSecureOperation("roles", "deleteRole", "Delete a role")
	opDelete.Parameters = []openapi3.ParameterOrRef{
		{Parameter: paramID},
	}

	success := s.deleteAPIResponses(&opDelete)
	handleError(s.reflector.Spec.AddOperation(routes.deleteRole.method, routes.deleteRole.getOAPI(), opDelete))

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		params := httprouter.ParamsFromContext(r.Context())
		id := params.ByName(paramID.Name)

		err := s.store.DeleteRole(ctx, &app.IDOrNameFilter{
			ID: id,
		})
		if err != nil {
			s.error(w, r, err)
			return
		}

		w.WriteHeader(success)
	}
}
