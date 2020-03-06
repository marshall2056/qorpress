package api

import (
	// "path/filepath"

	// "github.com/qorpress/qorpress/internal/assetfs"
	"github.com/qorpress/qorpress/internal/admin"
	"github.com/qorpress/qorpress/internal/qor"
	// "github.com/qorpress/qorpress/pkg/config"
	"github.com/qorpress/qorpress/pkg/config/application"
	"github.com/qorpress/qorpress/pkg/config/db"
	"github.com/qorpress/qorpress/pkg/models/posts"
	"github.com/qorpress/qorpress/pkg/models/users"
	// "github.com/qorpress/qorpress/pkg/config/bindatafs"
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

	API.AddResource(&posts.Post{})
	API.AddResource(&posts.Tag{})
	API.AddResource(&posts.Comment{})

	API.AddResource(&users.User{})

	API.AddResource(&posts.Category{})

	application.Router.Mount(app.Config.Prefix, API.NewServeMux(app.Config.Prefix))
}
