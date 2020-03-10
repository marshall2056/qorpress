package controllers

import (
	//"bytes"
	"fmt"
	//"html/template"
	"path/filepath"

	"github.com/qorpress/qorpress/core/admin"

	//"github.com/qorpress/qorpress/core/media"
	//"github.com/qorpress/qorpress/core/media/media_library"
	//"github.com/qorpress/qorpress/core/qor"
	"github.com/qorpress/qorpress-contrib/oniontree/models"
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
	themeDir := fmt.Sprintf(filepath.Join(config.Root, "plugins", "oniontree", "views"))
	controller := &Controller{View: render.New(&render.Config{
		AssetFileSystem: application.AssetFS.NameSpace("oniontree"),
	}, themeDir)}

	funcmapmaker.AddFuncMapMaker(controller.View)
	app.ConfigureAdmin(application.Admin)

	application.Router.Get("/onion-services", controller.Index)
	application.Router.Get("/onion-services/{code}", controller.Show)
	application.Router.Get("/onion-category/{code}", controller.Category)
	application.Router.Get("/onion-tag/{code}", controller.Tag)
}

// ConfigureAdmin configure admin interface
func (App) ConfigureAdmin(Admin *admin.Admin) {
	// OnionService Management
	Admin.AddMenu(&admin.Menu{Name: "OnionTree Management", Priority: 1})

	category := Admin.AddResource(&models.OnionCategory{}, &admin.Config{Menu: []string{"OnionTree Management"}, Priority: -3})
	category.Meta(&admin.Meta{Name: "Categories", Type: "select_many"})

	pks := Admin.AddResource(&models.OnionPublicKey{}, &admin.Config{Menu: []string{"OnionTree Management"}})
	pks.Meta(&admin.Meta{
		Name: "Value",
		Type: "text",
	})

	Admin.AddResource(&models.OnionLink{}, &admin.Config{Menu: []string{"OnionTree Management"}, Priority: -2})

	Admin.AddResource(&models.OnionTag{}, &admin.Config{Menu: []string{"OnionTree Management"}, Priority: -2})

	// Add OnionService
	srv := Admin.AddResource(&models.OnionService{}, &admin.Config{Menu: []string{"OnionTree Management"}})

	srvPropertiesRes := srv.Meta(&admin.Meta{Name: "ServiceProperties"}).Resource
	srvPropertiesRes.NewAttrs(&admin.Section{
		Rows: [][]string{{"Name", "Value"}},
	})
	srvPropertiesRes.EditAttrs(&admin.Section{
		Rows: [][]string{{"Name", "Value"}},
	})

	srv.Meta(&admin.Meta{
		Name: "Description",
		Type: "rich_editor",
	})

	srv.Meta(&admin.Meta{Name: "Category", Config: &admin.SelectOneConfig{AllowBlank: true}})
	// srv.Meta(&admin.Meta{Name: "Collections", Config: &admin.SelectManyConfig{SelectMode: "bottom_sheet"}})

	srv.Filter(&admin.Filter{
		Name: "Name",
		Type: "string",
	})

	srv.Filter(&admin.Filter{
		Name: "Code",
	})

	srv.Filter(&admin.Filter{
		Name: "CreatedAt",
	})

	type updateInfo struct {
		CategoryID uint
		Category   *models.OnionCategory
	}

	srv.SearchAttrs("Name", "Code", "Category.Name", "Tag.Name")
	srv.IndexAttrs("Name")
	
	srv.EditAttrs(
		&admin.Section{
			Title: "Seo Meta",
			Rows: [][]string{
				{"Seo"},
			}},
		&admin.Section{
			Title: "Basic Information",
			Rows: [][]string{
				{"Name"},
				{"Code"},
			}},
		&admin.Section{
			Title: "Organization",
			Rows: [][]string{
				{"Category"},
			}},
		"Tags",
		"ServiceProperties",
		"Description",
		"Links",
		"PublicKeys",
	)
	srv.ShowAttrs(srv.EditAttrs())
	srv.NewAttrs(srv.EditAttrs())

	srv.Action(&admin.Action{
		Name: "View On Site",
		URL: func(record interface{}, context *admin.Context) string {
			if srv, ok := record.(*models.OnionService); ok {
				return fmt.Sprintf("/onion-service/%v", srv.Code)
			}
			return "#"
		},
		Modes: []string{"menu_item", "edit"},
	})

}
