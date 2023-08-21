package sql

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/enverbisevac/go-project/app"
	"github.com/enverbisevac/go-project/app/sql/sqlite"
	"github.com/rs/zerolog/log"
)

func setupTest(t *testing.T) (*DB, func()) {
	t.Helper()
	dbtx, err := sqlite.New(":memory:")
	if err != nil {
		t.Fatalf("Error opening db, err: %v", err)
		return nil, func() {}
	}
	db, err := New(dbtx, true)
	if err != nil {
		t.Fatalf("error initializing db, err: %v", err)
		return nil, func() {}
	}
	return db, func() {
		db.Close()
	}
}

func TestSqliteLock(t *testing.T) {
	db, teardown := setupTest(t)
	defer teardown()

	wg := sync.WaitGroup{}

	// With transactions
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(counter int) {
			defer wg.Done()
			for k := 0; k < 100; k++ {
				role := app.RoleAggregate{
					Role: app.Role{
						Name: "Role" + fmt.Sprintf("%03d", counter) + fmt.Sprintf("%03d", k),
					},
				}
				err := db.AddRole(context.Background(), &role)
				if err != nil {
					t.Errorf("error inserting the role %s, err: %v", role.Name, app.SourceError(err))
					return
				}
			}
		}(i)
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(counter int) {
			defer wg.Done()
			var (
				role app.RoleAggregate
				err  error
			)
			for k := 0; k < 100; k++ {
				for {
					role, err = db.GetRole(context.Background(), &app.IDOrNameFilter{
						Name: "Role" + fmt.Sprintf("%03d", counter) + fmt.Sprintf("%03d", k),
					})
					if err != nil {
						continue // until role is found
					}
					break
				}

				if err != nil {
					t.Errorf("Failed to read %v, err: %v", role.Name, app.SourceError(err))
					return
				}
			}
		}(i)
	}

	// // Without transactions
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(counter int) {
			defer wg.Done()
			for k := 0; k < 100; k++ {
				role := app.Role{
					Name: "NoTxRole" + fmt.Sprintf("%03d", counter) + fmt.Sprintf("%03d", k),
				}
				err := db.InsertRole(context.Background(), &role)
				if err != nil {
					t.Errorf("error inserting the %s, err: %v", role.Name, app.SourceError(err))
					return
				}
			}
		}(i)
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(counter int) {
			defer wg.Done()
			var (
				role *app.Role
				err  error
			)
			for k := 0; k < 100; k++ {
				for {
					role, err = db.getRole(context.Background(), &app.IDOrNameFilter{
						Name: "NoTxRole" + fmt.Sprintf("%03d", counter) + fmt.Sprintf("%03d", k),
					})
					if err != nil {
						continue // until role is found
					}
					break
				}

				if err != nil {
					t.Errorf("Failed to read %s, err: %v", role.Name, app.SourceError(err))
					return
				}
			}
		}(i)
	}
	wg.Wait()
}

func Test1(t *testing.T) {
	db, teardown := setupTest(t)
	defer teardown()

	role := &app.Role{
		Name: "NoTxRole",
	}
	err := db.InsertRole(context.Background(), role)
	if err != nil {
		t.Errorf("error inserting the %s, err: %v", role.Name, app.SourceError(err))
		return
	}
	log.Info().Msgf("Role inserted: %s", role.Name)

	role, err = db.getRole(context.Background(), &app.IDOrNameFilter{
		Name: "NoTxRole",
	})
	if err != nil {
		t.Errorf("Failed to read %s, err: %v", role.Name, app.SourceError(err))
	}
	log.Info().Msgf("Role found: %s", role.Name)
}
