package controllers

import (
	"fmt"
	"path/filepath"

	"github.com/qorpress/qorpress-contrib/flickr/models"
	"github.com/qorpress/qorpress/core/admin"
	"github.com/qorpress/qorpress/core/render"
	"github.com/qorpress/qorpress/pkg/config"
	"github.com/qorpress/qorpress/pkg/config/application"
	"github.com/qorpress/qorpress/pkg/utils/funcmapmaker"
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
func (app App) ConfigureApplication(application *application.Application) {
	themeDir := fmt.Sprintf(filepath.Join(config.Root, "plugins", "flickr", "views"))
	controller := &Controller{View: render.New(&render.Config{
		AssetFileSystem: application.AssetFS.NameSpace("flickr"),
	}, themeDir)}

	funcmapmaker.AddFuncMapMaker(controller.View)
	app.ConfigureAdmin(application.Admin)

	application.Router.Get("/tweets", controller.Index)
}

// ConfigureAdmin configure admin interface
func (App) ConfigureAdmin(Admin *admin.Admin) {
	// OnionService Management
	Admin.AddMenu(&admin.Menu{Name: "Twitter Management", Priority: 1})

	// Add Setting
	Admin.AddResource(&models.TwitterSetting{}, &admin.Config{Name: "Twitter API", Menu: []string{"Twitter Management"}, Singleton: true, Priority: 1})
}


/*
import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gopress/internal/qorpress/pkg/services"
)

func GetFlickr(c *gin.Context) {
	payload, e := services.GetFlickrImages(9)
	services.GetFlickrAlbums()
	if e != nil {
		c.JSON(http.StatusInternalServerError, nil)
	} else {
		c.JSON(http.StatusOK, payload)
	}
}
*/