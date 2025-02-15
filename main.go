package main

import (
	"embed"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"github.com/rprtr258/impulse/internal/app"
)

//go:embed all:frontend/dist
var assets embed.FS

func run() error {
	fs := afero.NewBasePathFs(afero.NewOsFs(), "dist")
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	app, startup, close := app.New(fs)
	defer close()

	// Create application with options
	return wails.Run(&options.App{
		Title:  "impulse",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        startup,
		Bind: []interface{}{
			app,
		},
	})
}

func main() {
	if err := run(); err != nil {
		log.Fatal().Err(err).Msg("App stopped unexpectedly")
	}
}
