package api

// public service API

import (
	"context"

	"github.com/cenkalti/backoff/v4"
	"github.com/enverbisevac/go-project/app"
)

// Client interface for accessing remote services.
type Client interface {
	Login(ctx context.Context, input *app.Credentials) (string, error)
	CreateUser(ctx context.Context, input *app.UserAggregate) error
}

// API holds client connection for making api calls.
type API struct {
	client Client
}

var ExpBackoff = backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 3)

// Create a new API with client REST or GRPC.
func New(client Client) *API {
	return &API{
		client: client,
	}
}

// Login to the system using input param.
func (a *API) Login(ctx context.Context, input *app.Credentials) (token string, err error) {
	backoff.Retry(func() error {
		token, err = a.client.Login(ctx, input)
		if err != nil && app.ErrorStatus(err) == app.StatusInternal {
			return err
		}
		return backoff.Permanent(err)
	}, ExpBackoff)
	return token, nil
}

// CreateUser creates new user in the system.
func (a *API) CreateUser(ctx context.Context, input *app.UserAggregate) error {
	backoff.Retry(func() error {
		err := a.client.CreateUser(ctx, input)
		if err != nil && app.ErrorStatus(err) == app.StatusInternal {
			return err
		}
		return backoff.Permanent(err)
	}, ExpBackoff)
	return nil
}
