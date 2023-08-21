package http

import (
	"net/http"

	"github.com/enverbisevac/go-project/app"
)

func (s *Server) permissionsHandler() http.HandlerFunc {
	// define openapi operation
	opPermissions := createSecureOperation("permissions", "listPermissions", "List all permissions")

	statusCode := s.getAPIResponses(&opPermissions, app.Permissions)
	handleError(s.reflector.Spec.AddOperation(routes.permissions.method, routes.permissions.getOAPI(), opPermissions))

	return func(w http.ResponseWriter, r *http.Request) {
		err := JSON(w, statusCode, app.Permissions)
		if err != nil {
			s.error(w, r, err)
		}
	}
}
