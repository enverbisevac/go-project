package app

import (
	"context"
	"time"
)

type Session interface {
	UserID() string
}

type JWTClaims interface {
	Valid(t time.Time) bool
	AcceptAudience(stringOrURI string) bool
}

type AuthUser struct {
	ID   string
	Salt string
}

func (u AuthUser) UserID() string {
	return u.ID
}

type UserClaims struct {
	JWTClaims
	AuthUser
}

type JWTManager interface {
	Generate(user *AuthUser) ([]byte, time.Time, error)
	Verify(ctx context.Context, accessToken string) (*UserClaims, error)
}

type Mailer interface {
	SendError(err error, trace []byte)
}
