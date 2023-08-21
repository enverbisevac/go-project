package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/enverbisevac/go-project/app"
	"github.com/rs/zerolog/log"
)

func (s *Server) notFound(w http.ResponseWriter, r *http.Request) {
	message := "The requested resource could not be found"
	s.error(w, r, app.ErrNotFound(message))
}

func (s *Server) methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("The %s method is not supported for this resource", r.Method)
	s.error(w, r, app.ErrNotImplemented(message))
}

func (s *Server) invalidAuthToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	s.error(w, r, app.ErrUnauthorized("Invalid authentication token"))
}

func (s *Server) authRequired(w http.ResponseWriter, r *http.Request) {
	s.error(w, r, app.ErrUnauthenticated("You must be authenticated to access this resource"))
}

func (s *Server) authzRequired(w http.ResponseWriter, r *http.Request) {
	s.error(w, r, app.ErrUnauthorized("You must have permission to access this resource"))
}

func (s *Server) basicAuthRequired(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)

	message := "You must be authenticated to access this resource"
	s.error(w, r, app.ErrUnauthorized(message))
}

func (s *Server) invalidBody(w http.ResponseWriter, r *http.Request, contentType string, err error) {
	s.error(w, r, app.ErrInvalid("body contains badly-formed %s", contentType, err))
}

// Error prints & optionally logs an error message.
func (s *Server) error(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}
	// Extract error code & message.
	code, message, payload := app.ErrorStatus(err), app.ErrorMessage(err), app.ErrorPayload(err)

	// Log & report internal errors.
	if code == app.StatusInternal {
		fmterr := fmt.Errorf("%v, internal: %w", err, app.SourceError(err))
		log.Err(fmterr).Stack().Send()
	}

	// Print user message to response based on reqeust accept header.
	switch r.Header.Get("Accept") {
	case "application/json":
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(ErrorStatusCode(code))
		json.NewEncoder(w).Encode(&ErrorResponse{
			Error:   message,
			Payload: payload,
		})

	default:
		http.Error(w, message, ErrorStatusCode(code))
	}
}

// ErrorResponse represents a JSON structure for error output.
type ErrorResponse struct {
	Error   string `json:"error"`
	Payload any    `json:"payload,omitempty"`
}

// lookup of application error codes to HTTP status codes.
var codes = map[app.Status]int{
	app.StatusConflict:        http.StatusConflict,
	app.StatusInvalid:         http.StatusBadRequest,
	app.StatusNotFound:        http.StatusNotFound,
	app.StatusNotImplemented:  http.StatusNotImplemented,
	app.StatusUnauthenticated: http.StatusUnauthorized,
	app.StatusUnauthorized:    http.StatusForbidden,
	app.StatusInternal:        http.StatusInternalServerError,
}

// ErrorStatusCode returns the associated HTTP status code for a APP error code.
func ErrorStatusCode(code app.Status) int {
	if v, ok := codes[code]; ok {
		return v
	}
	return http.StatusInternalServerError
}

// FromErrorStatusCode returns the associated APP code for an HTTP status code.
func FromErrorStatusCode(code int) app.Status {
	for k, v := range codes {
		if v == code {
			return k
		}
	}
	return app.StatusInternal
}
