package http

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/enverbisevac/go-project/pkg/ptr"
	"github.com/swaggest/jsonschema-go"
	"github.com/swaggest/openapi-go/openapi3"
)

const (
	bearerToken = "bearerToken"
	apiKey      = "apiKey"
)

var paramID = createParam("id", openapi3.ParameterInPath, true, openapi3.SchemaTypeString)

func newReflector() *openapi3.Reflector {
	reflector := openapi3.Reflector{}
	reflector.Spec = &openapi3.Spec{
		Openapi: "3.0.3",
		Info: openapi3.Info{
			Title:   "ElasticPOS app API",
			Version: "1.0.0",
		},
		Servers: []openapi3.Server{
			{
				URL: "/",
			},
		},
		Components: &openapi3.Components{
			SecuritySchemes: &openapi3.ComponentsSecuritySchemes{
				MapOfSecuritySchemeOrRefValues: map[string]openapi3.SecuritySchemeOrRef{
					bearerToken: {
						SecurityScheme: &openapi3.SecurityScheme{
							HTTPSecurityScheme: &openapi3.HTTPSecurityScheme{
								Scheme:       "bearer",
								BearerFormat: ptr.From("JWT"),
							},
						},
					},
					apiKey: {
						SecurityScheme: &openapi3.SecurityScheme{
							APIKeySecurityScheme: &openapi3.APIKeySecurityScheme{
								Name: "API-Key",
								In:   openapi3.APIKeySecuritySchemeInHeader,
							},
						},
					},
				},
			},
		},
	}

	reflector.DefaultOptions = []func(r *jsonschema.ReflectContext){
		StripDefinitionName("app", "Http", "Aggregate", "Response"),
	}

	return &reflector
}

func StripDefinitionName(text ...string) func(rc *jsonschema.ReflectContext) {
	return func(rc *jsonschema.ReflectContext) {
		rc.DefName = func(t reflect.Type, defaultDefName string) string {
			s := defaultDefName
			for _, ps := range text {
				s = strings.TrimPrefix(s, ps)
				s = strings.ReplaceAll(s, "["+ps, "[")
				s = strings.TrimSuffix(s, ps)
				s = strings.ReplaceAll(s, ps+"]", "]")
			}

			return s
		}
	}
}

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

func createParam(name string, in openapi3.ParameterIn, required bool, t openapi3.SchemaType) *openapi3.Parameter {
	return &openapi3.Parameter{
		Name:     name,
		In:       in,
		Required: ptr.From(required),
		Schema: &openapi3.SchemaOrRef{
			Schema: &openapi3.Schema{
				Type: ptr.From(t),
			},
		},
	}
}

func createOperation(tag string, id string, summary string) openapi3.Operation {
	return openapi3.Operation{
		Tags:    []string{tag},
		ID:      ptr.From(id),
		Summary: ptr.From(summary),
	}
}

func createSecureOperation(tag string, id string, summary string) openapi3.Operation {
	op := createOperation(tag, id, summary)
	op.Security = []map[string][]string{
		{
			bearerToken: []string{},
			apiKey:      []string{},
		},
	}
	return op
}

func (s *Server) createAPIResponses(operation *openapi3.Operation, output any) int {
	const statusCode = http.StatusCreated
	handleError(s.reflector.SetJSONResponse(operation, output, statusCode))
	handleError(s.reflector.SetJSONResponse(operation, new(ErrorResponse), http.StatusBadRequest))
	handleError(s.reflector.SetJSONResponse(operation, new(ErrorResponse), http.StatusUnauthorized))
	handleError(s.reflector.SetJSONResponse(operation, new(ErrorResponse), http.StatusForbidden))
	handleError(s.reflector.SetJSONResponse(operation, new(ErrorResponse), http.StatusInternalServerError))
	return statusCode
}

func (s *Server) getAPIResponses(operation *openapi3.Operation, output any) int {
	const statusCode = http.StatusOK
	handleError(s.reflector.SetJSONResponse(operation, output, statusCode))
	handleError(s.reflector.SetJSONResponse(operation, new(ErrorResponse), http.StatusUnauthorized))
	handleError(s.reflector.SetJSONResponse(operation, new(ErrorResponse), http.StatusForbidden))
	handleError(s.reflector.SetJSONResponse(operation, new(ErrorResponse), http.StatusNotFound))
	handleError(s.reflector.SetJSONResponse(operation, new(ErrorResponse), http.StatusInternalServerError))
	return statusCode
}

func (s *Server) updateAPIResponses(operation *openapi3.Operation, output any) int {
	const statusCode = http.StatusOK
	handleError(s.reflector.SetJSONResponse(operation, output, statusCode))
	handleError(s.reflector.SetJSONResponse(operation, new(ErrorResponse), http.StatusBadRequest))
	handleError(s.reflector.SetJSONResponse(operation, new(ErrorResponse), http.StatusUnauthorized))
	handleError(s.reflector.SetJSONResponse(operation, new(ErrorResponse), http.StatusForbidden))
	handleError(s.reflector.SetJSONResponse(operation, new(ErrorResponse), http.StatusNotFound))
	handleError(s.reflector.SetJSONResponse(operation, new(ErrorResponse), http.StatusInternalServerError))
	return statusCode
}

func (s *Server) deleteAPIResponses(operation *openapi3.Operation) int {
	const statusCode = http.StatusNoContent
	handleError(s.reflector.SetJSONResponse(operation, nil, statusCode))
	handleError(s.reflector.SetJSONResponse(operation, new(ErrorResponse), http.StatusUnauthorized))
	handleError(s.reflector.SetJSONResponse(operation, new(ErrorResponse), http.StatusForbidden))
	handleError(s.reflector.SetJSONResponse(operation, new(ErrorResponse), http.StatusNotFound))
	handleError(s.reflector.SetJSONResponse(operation, new(ErrorResponse), http.StatusInternalServerError))
	return statusCode
}
