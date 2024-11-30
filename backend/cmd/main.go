package main

import (
	"os"
	"os/signal"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"

	"github.com/impulse-http/local-backend/internal/service"
)

func run() error {
	fs := afero.NewBasePathFs(afero.NewOsFs(), "dist")
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	addr := ":8090"
	log.Info().Str("addr", addr).Msg("Running")
	app, close := service.New(fs)
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt)
		<-ch
		close()
		app.Shutdown()
	}()
	return app.Listen(addr)
}

func main() {
	if err := run(); err != nil {
		log.Fatal().Err(err).Msg("App stopped unexpectedly")
	}
}
