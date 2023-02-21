package main

import (
	"context"
	"sda-filesystem/frontend"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

func main() {
	// Create an instance of the app structure
	logHandler := NewLogHandler()
	projectHandler := NewProjectHandler()
	app := NewApp(projectHandler)

	// Create application with options
	err := wails.Run(&options.App{
		Title: "Data Gateway",
		AssetServer: &assetserver.Options{
			Assets: frontend.Assets,
		},
		OnStartup: func(ctx context.Context) {
			logHandler.SetContext(ctx)
			projectHandler.SetContext(ctx)
			app.startup(ctx)
		},
		OnShutdown: app.shutdown,
		Bind: []interface{}{
			app,
			logHandler,
			projectHandler,
		},
		MinWidth:          800,
		Width:             800,
		Height:            575,
		HideWindowOnClose: true,
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
