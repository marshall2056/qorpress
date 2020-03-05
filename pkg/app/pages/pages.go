package pages

import (
	"fmt"

	"github.com/qorpress/admin"
	"github.com/qorpress/page_builder"
	"github.com/qorpress/qor"
	"github.com/qorpress/qor/resource"
	"github.com/qorpress/qor/utils"
	"github.com/qorpress/render"
	"github.com/qorpress/widget"

	adminapp "github.com/qorpress/qorpress/pkg/app/admin"
	"github.com/qorpress/qorpress/pkg/config/application"
	"github.com/qorpress/qorpress/pkg/config/db"
	"github.com/qorpress/qorpress/pkg/models/cms"
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
	controller := &Controller{View: render.New(&render.Config{
		AssetFileSystem: application.AssetFS.NameSpace("blog"),
	}, "pkg/app/pages/views")}

	funcmapmaker.AddFuncMapMaker(controller.View)
	app.ConfigureAdmin(application.Admin)
	application.Router.Get("/blog", controller.Index)
}

// ConfigureAdmin configure admin interface
func (App) ConfigureAdmin(Admin *admin.Admin) {
	Admin.AddMenu(&admin.Menu{Name: "Pages Management", Priority: 4})

	// Blog Management
	article := Admin.AddResource(&cms.Article{}, &admin.Config{Menu: []string{"Pages Management"}})
	article.IndexAttrs("ID", "VersionName", "ScheduledStartAt", "ScheduledEndAt", "Author", "Title")

	// Setup pages
	PageBuilderWidgets := widget.New(&widget.Config{DB: db.DB})
	PageBuilderWidgets.WidgetSettingResource = Admin.NewResource(&adminapp.QorWidgetSetting{}, &admin.Config{Name: "PageBuilderWidgets"})
	PageBuilderWidgets.WidgetSettingResource.NewAttrs(
		&admin.Section{
			Rows: [][]string{{"Kind"}, {"SerializableMeta"}},
		},
	)
	PageBuilderWidgets.WidgetSettingResource.AddProcessor(&resource.Processor{
		Handler: func(value interface{}, metaValues *resource.MetaValues, context *qor.Context) error {
			if widgetSetting, ok := value.(*adminapp.QorWidgetSetting); ok {
				if widgetSetting.Name == "" {
					var count int
					context.GetDB().Set(admin.DisableCompositePrimaryKeyMode, "off").Model(&adminapp.QorWidgetSetting{}).Count(&count)
					widgetSetting.Name = fmt.Sprintf("%v %v", utils.ToString(metaValues.Get("Kind").Value), count)
				}
			}
			return nil
		},
	})
	Admin.AddResource(PageBuilderWidgets, &admin.Config{Menu: []string{"Pages Management"}})

	page := page_builder.New(&page_builder.Config{
		Admin:       Admin,
		PageModel:   &cms.Page{},
		Containers:  PageBuilderWidgets,
		AdminConfig: &admin.Config{Name: "Pages", Menu: []string{"Pages Management"}, Priority: 1},
	})
	page.IndexAttrs("ID", "Title", "PublishLiveNow")
}
