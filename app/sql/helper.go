package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/enverbisevac/go-project/app"
	"github.com/jmoiron/sqlx"
)

func ToSliceOfAny[T any](s []T) []any {
	if len(s) == 0 {
		return nil
	}
	result := make([]any, len(s))
	for i, v := range s {
		result[i] = v
	}
	return result
}

type Modified interface {
	SetModified(value int64)
}

type Validator interface {
	Validate() error
}

type Object interface {
	Generator() (func() any, error)
	GetID() any
	SetID(id any)
	SetCreated(value int64)
}

func wrapError(err error, entity, condition string, args ...any) error {
	if err == nil {
		return nil
	}

	msg := fmt.Sprintf(condition, args...)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return app.ErrNotFound("%s not found with %s", entity, msg, err)
	default:
		return app.ErrInternal("failed to get %s data with %s", entity, msg, err)
	}
}

func insertSQL[T Object](
	ctx context.Context,
	dao DAO,
	query string,
	object T,
) error {
	if id := object.GetID(); id == nil || reflect.ValueOf(id).IsZero() {
		gen, err := object.Generator()
		if err != nil {
			return app.ErrInternal("generating id value failed!", err)
		}
		if gen != nil {
			value := gen()
			object.SetID(value)
		}
	}

	object.SetCreated(time.Now().Unix())

	query, args, err := sqlx.Named(query, object)
	if err != nil {
		return err
	}

	for _, arg := range args {
		if arg == nil {
			continue
		}
		v, ok := arg.(Validator)
		if ok && !reflect.ValueOf(v).IsZero() {
			if err = v.Validate(); err != nil {
				return app.ErrInvalid(err.Error(), err)
			}
		}
	}

	_, err = dao.ExecContext(ctx, query, args...)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return app.ErrConflict("row already exists", err)
		}
		if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			return app.ErrInvalid("FOREIGN KEY constraint failed", err)
		}
		return app.ErrInternal("failed to insert new row", err)
	}

	return nil
}

func getSQL[T any](
	ctx context.Context,
	dao DAO,
	query string,
	args ...any,
) (*T, error) {
	row := new(T)

	err := dao.GetContext(ctx, row, query, args...)
	if err != nil {
		return row, app.ErrInternal("failed to get rows", err)
	}
	return row, err
}

func querySQL[T any](
	ctx context.Context,
	dao DAO,
	query string,
	args ...any,
) ([]T, error) {
	rows := make([]T, 0, 100)

	err := dao.SelectContext(ctx, &rows, query, args...)
	if err != nil {
		return nil, app.ErrInternal("failed to get rows", err)
	}
	return rows, err
}

func updateSQL(
	ctx context.Context,
	dao DAO,
	query string,
	in any,
	params ...any,
) error {
	var (
		args []any
		err  error
	)

	value := reflect.ValueOf(in)
	if value.Kind() == reflect.Pointer && !value.IsNil() {
		value = value.Elem()
	}

	switch value.Kind() {
	case reflect.Struct:
		m, ok := in.(Modified)
		if ok {
			m.SetModified(time.Now().Unix())
		}
		query, args, err = sqlx.Named(query, in)
		if err != nil {
			return app.ErrInternal("failed to bind row fields", err)
		}
	case reflect.Map:
		query, args, err = sqlx.Named(query, in)
		if err != nil {
			return app.ErrInternal("failed to bind row fields", err)
		}
	case reflect.Slice:
		args = in.([]any)
	default:
		args = []any{in}
	}

	for _, arg := range args {
		v, ok := arg.(Validator)
		if ok && !reflect.ValueOf(v).IsZero() {
			if err = v.Validate(); err != nil {
				return app.ErrInvalid(err.Error(), err)
			}
		}
	}

	result, err := dao.ExecContext(ctx, query, append(args, params...)...)
	if err != nil {
		return app.ErrInternal("updating failed!", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return app.ErrInternal("update rows affected error", err)
	}

	if n == 0 {
		return app.ErrNotFound("not found")
	}

	return nil
}

func deleteSQL(
	ctx context.Context,
	dao DAO,
	query string,
	args ...any,
) error {
	result, err := dao.ExecContext(ctx, query, args...)
	if err != nil {
		return app.ErrInternal("failed to delete a row", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return app.ErrInternal("delete rows affected error", err)
	}

	if n == 0 {
		return app.ErrNotFound("rows not found")
	}

	return nil
}
