package home

import (
	"path/filepath"

	"github.com/gopress/internal/render"

	"github.com/gopress/qorpress/pkg/config"
	"github.com/gopress/qorpress/pkg/config/application"
	"github.com/gopress/qorpress/pkg/utils/funcmapmaker"
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
	themeDir := fmt.Sprintf(filepath.Join(config.Root, "themes", "qorpress", "views", "home"))
	controller := &Controller{View: render.New(&render.Config{
		AssetFileSystem: application.AssetFS.NameSpace("home"),
	}, themeDir)}

	funcmapmaker.AddFuncMapMaker(controller.View)
	application.Router.Get("/", controller.Index)
	application.Router.Get("/switch_locale", controller.SwitchLocale)
}
