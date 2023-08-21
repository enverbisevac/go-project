package app

import (
	"fmt"

	"github.com/enverbisevac/go-project/pkg/validator"
)

type Email string

func (e Email) Validate() error {
	if e == "" {
		return ErrEmailFieldIsRequired
	}
	if !validator.IsEmail(e) {
		return fmt.Errorf("'%s' %w", e, ErrEmailIsNotValid)
	}
	return nil
}

func (e Email) String() string {
	return string(e)
}

type Password string

func (p Password) Validate() error {
	if p == "" {
		return ErrPasswordIsRequired
	}
	if len(p) < 8 {
		return ErrPasswordIsTooShort
	}
	if len(p) >= 72 {
		return ErrPasswordIsTooLong
	}
	if validator.In(p.String(), validator.CommonPasswords...) {
		return ErrPasswordIsCommon
	}
	return nil
}

func (p Password) String() string {
	return string(p)
}
