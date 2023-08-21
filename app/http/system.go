package http

import (
	"net/http"
)

func (s *Server) status() http.HandlerFunc {
	opStatus := createOperation("system", "status", "System status details")

	handleError(s.reflector.SetRequest(&opStatus, nil, routes.status.method))
	handleError(s.reflector.SetJSONResponse(&opStatus, new(map[string]any), http.StatusOK))
	handleError(s.reflector.SetJSONResponse(&opStatus, new(ErrorResponse), http.StatusUnauthorized))
	handleError(s.reflector.SetJSONResponse(&opStatus, new(ErrorResponse), http.StatusInternalServerError))
	handleError(s.reflector.Spec.AddOperation(routes.status.method, routes.status.getOAPI(), opStatus))

	return func(w http.ResponseWriter, r *http.Request) {
		data := map[string]string{
			"Status": "OK",
		}

		err := JSON(w, http.StatusOK, data)
		if err != nil {
			s.error(w, r, err)
		}
	}
}
