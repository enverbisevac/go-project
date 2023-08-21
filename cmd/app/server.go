package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/enverbisevac/go-project/app"
	"github.com/enverbisevac/go-project/app/http"
	"github.com/enverbisevac/go-project/app/jwt"
	"github.com/enverbisevac/go-project/app/sql"
	"github.com/enverbisevac/go-project/app/sql/sqlite"
	"github.com/jxskiss/mcli"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func serverCmd() error {
	var flags struct {
		BaseURL string `cli:"--base-url    Base url for application" default:"http://localhost"`
		Port    int    `cli:"-p, --port    Port to listen on for HTTP requests" default:"4444"`
		Detach  bool   `cli:"-d, --detach  Detach process"`
		DSN     string `cli:"--dsn         Data source name"`
		Migrate bool   `cli:"--migrate     Run auto migration" default:"true"`
		LogDir  string `cli:"--log-dir     Set log dir"`
	}

	_, err := mcli.Parse(&flags)
	if err != nil {
		return err
	}

	// detach application
	if flags.Detach {
		args := make([]string, 0, len(os.Args))
		for _, arg := range os.Args[1:] {
			if arg != "-d" && arg != "--detach" {
				args = append(args, arg)
			}
		}
		cmd := exec.Command(os.Args[0], args...)
		cmd.Start()
		fmt.Println("[PID]", cmd.Process.Pid)
		os.Exit(0)
	}

	// setup logs
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	log.Info().Msg("Application started")

	if flags.DSN == "" {
		flags.DSN = "./app.db"
	}
	sqlt, err := sqlite.New(flags.DSN)
	if err != nil {
		return err
	}
	db, err := sql.New(sqlt, flags.Migrate)
	if err != nil {
		return err
	}
	defer db.Close()

	db.AddUser(context.Background(), &app.UserAggregate{
		User: app.User{
			Active:     true,
			FullName:   "Admin",
			IsAdmin:    true,
			DateJoined: time.Now().UnixMilli(),
			Email:      app.Email("admin@domain.com"),
			Password:   app.Password("SomePassword"),
		},
	})

	// initialize services
	jwtService := jwt.NewManager(flags.BaseURL, 24*time.Hour, db)
	httpService := http.New(http.Config{
		BaseURL: flags.BaseURL,
		Port:    flags.Port,
	}, jwtService, db, db, db)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := httpService.Start(); err != nil {
			log.Fatal().Msgf("error while starting http server, err: %v", err)
		}
	}()

	<-done
	log.Print("Server Stopped")

	if err := httpService.Stop(); err != nil {
		log.Fatal().Msgf("Server Shutdown Failed:%+v", err)
	}

	log.Info().Msg("Server stopped properly")

	return nil
}
