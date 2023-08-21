package http

import (
	"context"

	"github.com/enverbisevac/go-project/app"
	"github.com/enverbisevac/libs/httputil"
)

type HTTP interface {
	Get(ctx context.Context, rawurl string, out interface{}, options ...httputil.RequestOption) error
	Post(ctx context.Context, rawurl string, in, out interface{}, options ...httputil.RequestOption) error
	Patch(ctx context.Context, rawurl string, in, out interface{}, options ...httputil.RequestOption) error
	Delete(ctx context.Context, rawurl string, options ...httputil.RequestOption) error
}

type Client struct {
	client HTTP
	token  string
}

func NewClient(client HTTP) *Client {
	return &Client{
		client: client,
	}
}

func (r *Client) Login(ctx context.Context, input *app.Credentials) (string, error) {
	res := app.TokenData{}

	if err := input.Validate(); err != nil {
		return "", err
	}

	if err := r.client.Post(ctx, routes.login.path, input, &res); err != nil {
		return "", err
	}

	r.token = res.Token

	return res.Token, nil
}

func (r *Client) CreateUser(ctx context.Context, input *app.UserAggregate) error {
	res := &app.UserAggregate{}

	if err := input.Validate(); err != nil {
		return err
	}

	if err := r.client.Post(ctx, routes.createUser.path, input, res,
		httputil.WithAuthHeader("Bearer "+r.token)); err != nil {
		return err
	}

	return nil
}
