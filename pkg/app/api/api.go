package api

import (
	"github.com/qorpress/qorpress/core/admin"
	"github.com/qorpress/qorpress/core/qor"
	"github.com/qorpress/qorpress/pkg/config/application"
	"github.com/qorpress/qorpress/pkg/config/db"
	"github.com/qorpress/qorpress/pkg/models/posts"
	"github.com/qorpress/qorpress/pkg/models/users"
)

// New new home app
func New(config *Config) *App {
	if config.Prefix == "" {
		config.Prefix = "/api"
	}
	return &App{Config: config}
}

// App home app
type App struct {
	Config *Config
}

// Config home config struct
type Config struct {
	Prefix string
}

// ConfigureApplication configure application
func (app App) ConfigureApplication(application *application.Application) {

	API := admin.New(&qor.Config{
		DB: db.DB,
	})

	// how to generate swagger doc from that ?
	API.AddResource(&posts.Post{})
	API.AddResource(&posts.Tag{})
	API.AddResource(&posts.Comment{})
	API.AddResource(&users.User{})
	API.AddResource(&posts.Category{})

	// to do: iterate through plugins to register new api endpoints
	// for _, pluginRes := plug.Plugins
	// API.AddResource(pluginRes)

	application.Router.Mount(app.Config.Prefix, API.NewServeMux(app.Config.Prefix))
}
