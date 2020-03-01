package home

import (
	"github.com/qorpress/render"

	"github.com/qorpress/qorpress-example/pkg/config/application"
	"github.com/qorpress/qorpress-example/pkg/utils/funcmapmaker"
)

// New new home app
func New(config *Config) *App {
	return &App{Config: config}
}

// App home app
type App struct {
	Config *Config
}

// Config home config struct
type Config struct {
}

// ConfigureApplication configure application
func (App) ConfigureApplication(application *application.Application) {
	controller := &Controller{View: render.New(&render.Config{
		AssetFileSystem: application.AssetFS.NameSpace("home"),
	}, "pkg/app/home/views")}

	funcmapmaker.AddFuncMapMaker(controller.View)
	application.Router.Get("/", controller.Index)
	application.Router.Get("/switch_locale", controller.SwitchLocale)
}
