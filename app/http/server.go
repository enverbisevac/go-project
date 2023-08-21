package http

import (
	"context"
	"errors"
	"fmt"
	stdlog "log"
	"net/http"
	"time"

	"github.com/enverbisevac/go-project/app"
	"github.com/rs/zerolog/log"
	"github.com/swaggest/openapi-go/openapi3"
)

const (
	defaultIdleTimeout  = time.Minute
	defaultReadTimeout  = 10 * time.Second
	defaultWriteTimeout = 30 * time.Second

	defaultShutdownPeriod = 20 * time.Second
)

type Config struct {
	BaseURL string
	Port    int
}

type Server struct {
	http          *http.Server
	config        Config
	jwt           app.JWTManager
	authenticator app.Authenticator
	authorizer    app.Authorizer
	store         app.Storage
	reflector     *openapi3.Reflector
}

func New(config Config,
	jwt app.JWTManager,
	authenticator app.Authenticator,
	authorizer app.Authorizer,
	store app.Storage,
) *Server {
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Port),
		ErrorLog:     stdlog.New(log.Logger, "", 0),
		IdleTimeout:  defaultIdleTimeout,
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
	}

	server := &Server{
		http:          httpServer,
		config:        config,
		jwt:           jwt,
		authenticator: authenticator,
		authorizer:    authorizer,
		store:         store,
		reflector:     newReflector(),
	}

	httpServer.Handler = server.routes()

	return server
}

func (s *Server) Start() error {
	log.Info().Str("address", s.http.Addr).Msg("starting http server")
	err := s.http.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownPeriod)
	defer cancel()
	log.Info().Str("address", s.http.Addr).Msg("stopping http server")
	s.http.Shutdown(ctx)
	log.Info().Msg("http server stopped")

	return nil
}
