package posts

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"

	// "github.com/jinzhu/gorm"
	"github.com/qorpress/qorpress/core/admin"
	"github.com/qorpress/qorpress/core/media"
	"github.com/qorpress/qorpress/core/media/media_library"
	"github.com/qorpress/qorpress/core/qor"
	"github.com/qorpress/qorpress/core/render"
	"github.com/qorpress/qorpress/pkg/config"
	"github.com/qorpress/qorpress/pkg/config/application"
	"github.com/qorpress/qorpress/pkg/models/posts"
	"github.com/qorpress/qorpress/pkg/utils/funcmapmaker"
)

// var Genders = []string{"Men", "Women", "Kids"}

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
	themeDir := fmt.Sprintf(filepath.Join(config.Root, "themes", "qorpress", "views", "posts"))
	controller := &Controller{View: render.New(&render.Config{
		AssetFileSystem: application.AssetFS.NameSpace("posts"),
	}, themeDir)}

	funcmapmaker.AddFuncMapMaker(controller.View)
	app.ConfigureAdmin(application.Admin)

	application.Router.Get("/posts", controller.Index)
	application.Router.Get("/posts/{code}", controller.Show)
	application.Router.Get("/category/{code}", controller.Category)
	application.Router.Get("/tag/{code}", controller.Tag)
}

// ConfigureAdmin configure admin interface
func (App) ConfigureAdmin(Admin *admin.Admin) {
	// Post Management
	Admin.AddMenu(&admin.Menu{Name: "Post Management", Priority: 1})

	category := Admin.AddResource(&posts.Category{}, &admin.Config{Menu: []string{"Post Management"}, Priority: -3})
	category.Meta(&admin.Meta{Name: "Categories", Type: "select_many"})

	collection := Admin.AddResource(&posts.Collection{}, &admin.Config{Menu: []string{"Post Management"}, Priority: -2})

	Admin.AddResource(&posts.Link{}, &admin.Config{Menu: []string{"Post Management"}, Priority: -2})

	Admin.AddResource(&posts.Comment{}, &admin.Config{Menu: []string{"Post Management"}, Priority: -2})

	Admin.AddResource(&posts.Tag{}, &admin.Config{Menu: []string{"Post Management"}, Priority: -2})

	// Add PostImage as Media Libraray
	PostImagesResource := Admin.AddResource(&posts.PostImage{}, &admin.Config{Menu: []string{"Post Management"}, Priority: -1})

	PostImagesResource.Filter(&admin.Filter{
		Name:       "SelectedType",
		Label:      "Media Type",
		Operations: []string{"contains"},
		Config:     &admin.SelectOneConfig{Collection: [][]string{{"video", "Video"}, {"image", "Image"}, {"file", "File"}, {"video_link", "Video Link"}}},
	})
	PostImagesResource.Filter(&admin.Filter{
		Name:   "Category",
		Config: &admin.SelectOneConfig{RemoteDataResource: category},
	})
	PostImagesResource.IndexAttrs("File", "Title")

	// Add Post
	post := Admin.AddResource(&posts.Post{}, &admin.Config{Menu: []string{"Post Management"}})

	postPropertiesRes := post.Meta(&admin.Meta{Name: "PostProperties"}).Resource
	postPropertiesRes.NewAttrs(&admin.Section{
		Rows: [][]string{{"Name", "Value"}},
	})
	postPropertiesRes.EditAttrs(&admin.Section{
		Rows: [][]string{{"Name", "Value"}},
	})

	post.Meta(&admin.Meta{Name: "Description", Config: &admin.RichEditorConfig{Plugins: []admin.RedactorPlugin{
		{Name: "medialibrary", Source: "/admin/assets/javascripts/qor_redactor_medialibrary.js"},
		{Name: "table", Source: "/vendors/redactor_table.js"},
	},
		Settings: map[string]interface{}{
			"medialibraryUrl": "/admin/post_images",
		},
	}})
	post.Meta(&admin.Meta{Name: "Category", Config: &admin.SelectOneConfig{AllowBlank: true}})
	post.Meta(&admin.Meta{Name: "Collections", Config: &admin.SelectManyConfig{SelectMode: "bottom_sheet"}})

	post.Meta(&admin.Meta{Name: "MainImage", Config: &media_library.MediaBoxConfig{
		RemoteDataResource: PostImagesResource,
		Max:                1,
		Sizes: map[string]*media.Size{
			"original": {Width: 560, Height: 700},
		},
	}})

	post.Meta(&admin.Meta{Name: "MainImageURL", Valuer: func(record interface{}, context *qor.Context) interface{} {
		if p, ok := record.(*posts.Post); ok {
			result := bytes.NewBufferString("")
			tmpl, _ := template.New("").Parse("<img src='{{.image}}'></img>")
			tmpl.Execute(result, map[string]string{"image": p.MainImageURL()})
			return template.HTML(result.String())
		}
		return ""
	}})

	post.Filter(&admin.Filter{
		Name:   "Collections",
		Config: &admin.SelectOneConfig{RemoteDataResource: collection},
	})

	post.Filter(&admin.Filter{
		Name: "Featured",
	})

	post.Filter(&admin.Filter{
		Name: "Name",
		Type: "string",
	})

	post.Filter(&admin.Filter{
		Name: "Code",
	})

	post.Filter(&admin.Filter{
		Name: "CreatedAt",
	})

	post.Action(&admin.Action{
		Name:        "Import Post",
		URLOpenType: "slideout",
		URL: func(record interface{}, context *admin.Context) string {
			return "/admin/workers/new?job=Import Posts"
		},
		Modes: []string{"collection"},
	})

	type updateInfo struct {
		CategoryID uint
		Category   *posts.Category
	}

	updateInfoRes := Admin.NewResource(&updateInfo{})
	post.Action(&admin.Action{
		Name:     "Update Info",
		Resource: updateInfoRes,
		Handler: func(argument *admin.ActionArgument) error {
			newPostInfo := argument.Argument.(*updateInfo)
			for _, record := range argument.FindSelectedRecords() {
				fmt.Printf("%#v\n", record)
				if post, ok := record.(*posts.Post); ok {
					if newPostInfo.Category != nil {
						post.Category = *newPostInfo.Category
					}
					argument.Context.GetDB().Save(post)
				}
			}
			return nil
		},
		Modes: []string{"batch"},
	})

	post.UseTheme("grid")

	post.SearchAttrs("Name", "Code", "Category.Name", "Tag.Name")
	post.IndexAttrs("MainImageURL", "Name", "Featured", "VersionName", "PublishLiveNow")
	post.EditAttrs(
		&admin.Section{
			Title: "Seo Meta",
			Rows: [][]string{
				{"Seo"},
			}},
		&admin.Section{
			Title: "Basic Information",
			Rows: [][]string{
				{"Name", "Featured"},
				{"Code"},
				{"MainImage"},
			}},
		&admin.Section{
			Title: "Organization",
			Rows: [][]string{
				{"Category"},
				{"Collections"},
			}},
		"Tags",
		"PostProperties",
		"Description",
		"PublishReady",
		"Images",
		"Links",
		"Comments",
	)
	post.ShowAttrs(post.EditAttrs())
	post.NewAttrs(post.EditAttrs())

	post.Action(&admin.Action{
		Name: "View On Site",
		URL: func(record interface{}, context *admin.Context) string {
			if post, ok := record.(*posts.Post); ok {
				return fmt.Sprintf("/posts/%v", post.Code)
			}
			return "#"
		},
		Modes: []string{"menu_item", "edit"},
	})

}
