package http

import (
	"net/http"
	"time"

	"github.com/enverbisevac/go-project/app"
	"github.com/enverbisevac/libs/httputil"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog/hlog"
	"github.com/swaggest/swgui"
	"github.com/swaggest/swgui/v4emb"
)

const (
	paramEmpty = ""
)

var routes = struct {
	status         route
	login          route
	createUser     route
	getUser        route
	updateUser     route
	updatePassword route
	deleteUser     route
	permissions    route
	createRole     route
	getRole        route
	updateRole     route
	deleteRole     route
}{
	status:         route{path: "/status", method: http.MethodGet},
	login:          route{path: "/login", method: http.MethodPost},
	createUser:     route{path: "/users", method: http.MethodPost},
	getUser:        route{path: "/users/:id", method: http.MethodGet},
	updateUser:     route{path: "/users/:id", method: http.MethodPut},
	updatePassword: route{path: "/users/:id/password", method: http.MethodPut},
	deleteUser:     route{path: "/users/:id", method: http.MethodDelete},
	permissions:    route{path: "/permissions", method: http.MethodGet},
	createRole:     route{path: "/roles", method: http.MethodPost},
	getRole:        route{path: "/roles/:id", method: http.MethodGet},
	updateRole:     route{path: "/roles/:id", method: http.MethodPut},
	deleteRole:     route{path: "/roles/:id", method: http.MethodDelete},
}

func (s *Server) openapi(w http.ResponseWriter, r *http.Request) {
	schema, err := s.reflector.Spec.MarshalYAML()
	if err != nil {
		s.error(w, r, err)
		return
	}

	w.Write(schema)
}

func (s *Server) routes() http.Handler {
	mux := httprouter.New()

	mux.NotFound = http.HandlerFunc(s.notFound)
	mux.MethodNotAllowed = http.HandlerFunc(s.methodNotAllowed)

	// swagger
	mux.HandlerFunc(http.MethodGet, "/openapi.yaml", s.openapi)
	mux.Handler(http.MethodGet, "/openapi/*swagger", v4emb.NewHandlerWithConfig(swgui.Config{
		Title:       "Elastic POS - app API",
		SwaggerJSON: "/openapi.yaml",
		BasePath:    "/openapi/swagger",
		SettingsUI: map[string]string{
			"defaultModelsExpandDepth": "1",
		},
	}))

	// API routes

	// system
	mux.HandlerFunc("GET", "/status", s.status())

	// auth
	mux.HandlerFunc(routes.login.method, routes.login.path, s.loginHandler())
	mux.HandlerFunc(routes.permissions.method, routes.permissions.path, s.permissionsHandler())

	// users
	mux.Handler(routes.createUser.method, routes.createUser.path, s.authorize(
		s.requireAuthUser(s.createUserHandler()),
		app.PermissionCreateUser, paramEmpty),
	)
	mux.Handler(routes.updateUser.method, routes.updateUser.path,
		s.authorize(s.requireAuthUser(s.updateUserHandler()),
			app.PermissionUpdateUser, paramID.Name),
	)
	mux.Handler(routes.updatePassword.method, routes.updatePassword.path, s.requireAuthUser(
		s.updateUserPasswordHandler()),
	)
	mux.Handler(routes.getUser.method, routes.getUser.path, s.authorize(
		s.requireAuthUser(s.getUserHandler()),
		app.PermissionViewUser, paramID.Name),
	)
	mux.Handler(routes.deleteUser.method, routes.deleteUser.path, s.authorize(
		s.requireAuthUser(s.deleteUserHandler()),
		app.PermissionDeleteUser, paramID.Name),
	)

	// roles
	mux.Handler(routes.status.method, routes.createRole.path, s.authorize(
		s.requireAuthUser(http.HandlerFunc(s.createRoleHandler())),
		app.PermissionCreateRole, paramEmpty),
	)
	mux.Handler(routes.getRole.method, routes.getRole.path, s.authorize(
		s.requireAuthUser(http.HandlerFunc(s.getRoleHandler())),
		app.PermissionViewRole, paramID.Name),
	)
	mux.Handler(routes.updateRole.method, routes.updateRole.path, s.authorize(
		s.requireAuthUser(http.HandlerFunc(s.updateRoleHandler())),
		app.PermissionUpdateRole, paramID.Name),
	)
	mux.Handler(routes.deleteRole.method, routes.deleteRole.path, s.authorize(
		s.requireAuthUser(http.HandlerFunc(s.deleteRoleHandler())),
		app.PermissionDeleteRole, paramID.Name),
	)

	// Web routes

	mux.Handler("GET", "/protected", s.requireAuthUser(
		http.HandlerFunc(s.protected)),
	)
	mux.Handler("GET", "/basic-auth-protected", s.requireBasicAuth(
		http.HandlerFunc(s.protected)),
	)
	mux.HandlerFunc("GET", "/", s.index)

	c := httputil.NewChain(
		s.loggerHandler,
		hlog.RequestIDHandler("requestId", "X-Request-Id"),
		hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
			hlog.FromRequest(r).Info().
				Str("method", r.Method).
				Stringer("url", r.URL).
				Int("status", status).
				Int("size", size).
				Dur("duration", duration).
				Send()
		}),
		s.authenticate,
		s.recoverPanic,
	)

	return c.Then(mux)
}

type route struct {
	path   string
	method string
}

func (r route) getOAPI() string {
	var (
		s     string
		found bool
	)
	for _, v := range r.path {
		if v == ':' {
			s += "{"
			found = true
		} else if found && v == '/' {
			s += "}/"
			found = false
		} else {
			s += string(v)
		}
	}
	if found {
		s += "}"
	}
	return s
}
