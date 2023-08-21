package jwt

import (
	"context"
	"strings"
	"time"

	"github.com/enverbisevac/go-project/app"

	"github.com/pascaldekloe/jwt"
)

type Manager struct {
	baseURL  string
	duration time.Duration
	store    app.Storage
}

func NewManager(baseURL string, duration time.Duration, store app.Storage) *Manager {
	return &Manager{
		baseURL:  baseURL,
		duration: duration,
		store:    store,
	}
}

func (m *Manager) Generate(user *app.AuthUser) ([]byte, time.Time, error) {
	var claims jwt.Claims
	claims.Subject = user.ID

	expiry := time.Now().Add(m.duration)
	claims.Issued = jwt.NewNumericTime(time.Now())
	claims.NotBefore = jwt.NewNumericTime(time.Now())
	claims.Expires = jwt.NewNumericTime(expiry)

	claims.Issuer = m.baseURL
	claims.Audiences = []string{m.baseURL}

	jwtBytes, err := claims.HMACSign(jwt.HS256, []byte(user.Salt))
	if err != nil {
		return nil, time.Time{}, err
	}

	return jwtBytes, expiry, nil
}

func (m *Manager) Verify(ctx context.Context, token string) (*app.UserClaims, error) {
	if strings.HasPrefix(token, "Bearer") {
		_, after, found := strings.Cut(token, " ")
		if found {
			token = after
		}
	}
	claims, err := jwt.ParseWithoutCheck([]byte(token))
	if err != nil {
		return nil, app.ErrInternal("failed to parse token", err)
	}

	user, err := m.store.GetUser(ctx, app.UserFilter{
		ID: claims.Subject,
	})
	if err != nil {
		return nil, app.ErrUnauthorized("unauthorized: %w", err)
	}

	claims, err = jwt.HMACCheck([]byte(token), []byte(user.Salt))
	if err != nil {
		return nil, err
	}

	if !claims.Valid(time.Now()) {
		return nil, app.ErrUnauthorized("token is not valid")
	}

	if claims.Issuer != m.baseURL {
		return nil, app.ErrUnauthorized("issuer %s is not valid", m.baseURL)
	}

	if !claims.AcceptAudience(m.baseURL) {
		return nil, app.ErrUnauthorized("audience %s mismatch", m.baseURL)
	}

	return &app.UserClaims{
		JWTClaims: claims,
		AuthUser: app.AuthUser{
			ID: user.ID,
		},
	}, nil
}
