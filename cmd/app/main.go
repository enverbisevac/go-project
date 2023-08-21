package main

import (
	"fmt"

	"github.com/jxskiss/mcli"
	"github.com/rs/zerolog/log"

	"github.com/enverbisevac/go-project/version"
)

func main() {
	mcli.Add("server", func() {
		if err := serverCmd(); err != nil {
			log.Fatal().Err(err).Msg("Error while running the http services")
		}
	}, "Run http server")
	mcli.Add("version", func() {
		fmt.Printf("version: %s\n", version.Get())
	}, "Show app version")
	mcli.Run()
}
