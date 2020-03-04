package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"runtime"
	"time"
	"strings"
	"strconv"
	"bytes"

	// "github.com/qorpress/wildcard_router"
	"github.com/biezhi/gorm-paginator/pagination"
	"github.com/qorpress/media/oss"
	"github.com/qorpress/worker"
	"github.com/qorpress/i18n/exchange_actions"
	"github.com/qorpress/exchange/backends/csv"
	"github.com/qorpress/exchange"
	"github.com/qorpress/qor/resource"
	"github.com/qorpress/page_builder"
	"github.com/qorpress/widget"
	"github.com/Masterminds/sprig"
	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/koreset/gtf"
	"github.com/qorpress/admin"
	"github.com/qorpress/assetfs"
	"github.com/qorpress/auth"
	"github.com/qorpress/auth/providers/facebook"
	"github.com/qorpress/auth/providers/github"
	"github.com/qorpress/auth/providers/google"
	"github.com/qorpress/auth/providers/password"
	"github.com/qorpress/auth/providers/twitter"
	"github.com/qorpress/auth_themes/clean"
	"github.com/qorpress/i18n"
	"github.com/qorpress/i18n/backends/database"
	i18n_database "github.com/qorpress/i18n/backends/database"
	"github.com/qorpress/i18n/backends/yaml"
	"github.com/qorpress/l10n"
	"github.com/qorpress/media"
	"github.com/qorpress/media/asset_manager"
	"github.com/qorpress/media/media_library"
	"github.com/qorpress/publish2"
	"github.com/qorpress/qor"
	qor_utils "github.com/qorpress/qor/utils"
	"github.com/qorpress/redirect_back"
	qor_seo "github.com/qorpress/seo"
	"github.com/qorpress/session/manager"
	"github.com/qorpress/sorting"
	"github.com/qorpress/validations"
	uuid "github.com/satori/go.uuid"
	stats "github.com/semihalev/gin-stats"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/swaggo/gin-swagger/example/basic/docs"
	"golang.org/x/crypto/acme/autocert"

	_ "github.com/qorpress/qorpress/config/bindatafs"
	"github.com/qorpress/qorpress/pkg/config"
	"github.com/qorpress/qorpress/pkg/config/bindatafs"
	"github.com/qorpress/qorpress/pkg/controllers"
	"github.com/qorpress/qorpress/pkg/models"
	"github.com/qorpress/qorpress/pkg/services"
	"github.com/qorpress/qorpress/pkg/utils"
)

var db *gorm.DB
var I18n *i18n.I18n
var funcMaps template.FuncMap
var templates *template.Template
var Auth *auth.Auth
var RedirectBack *redirect_back.RedirectBack

// AutoMigrate run auto migration
func AutoMigrate(values ...interface{}) {
	for _, value := range values {
		db.AutoMigrate(value)
	}
}

func SetupAuth() {

	RedirectBack = redirect_back.New(&redirect_back.Config{
		SessionManager:  manager.SessionManager,
		IgnoredPrefixes: []string{"/auth"},
	})

	Auth = clean.New(&auth.Config{
		DB: db,
		// NO NEED TO CONFIG RENDER, AS IT'S CONFIGED IN CLEAN THEME
		// Render:     render.New(&render.Config{AssetFileSystem: bindatafs.AssetFS.NameSpace("auth")}),
		Mailer:     config.Mailer,
		UserModel:  models.User{},
		Redirector: auth.Redirector{RedirectBack},
	})

	// Register Auth providers
	// Allow use username/password
	Auth.RegisterProvider(password.New(&password.Config{}))

	// Allow use Github
	Auth.RegisterProvider(github.New(&github.Config{
		ClientID:     config.Config.Auth.Github.ClientID,
		ClientSecret: config.Config.Auth.Github.ClientSecret,
	}))

	// Allow use Google
	Auth.RegisterProvider(google.New(&google.Config{
		ClientID:       config.Config.Auth.Google.ClientID,
		ClientSecret:   config.Config.Auth.Google.ClientSecret,
		AllowedDomains: []string{}, // Accept all domains, instead you can pass a whitelist of acceptable domains
	}))

	// Allow use Facebook
	Auth.RegisterProvider(facebook.New(&facebook.Config{
		ClientID:     config.Config.Auth.Facebook.ClientID,
		ClientSecret: config.Config.Auth.Facebook.ClientSecret,
	}))

	// Allow use Twitter
	Auth.RegisterProvider(twitter.New(&twitter.Config{
		ClientID:     config.Config.Auth.Twitter.ClientID,
		ClientSecret: config.Config.Auth.Twitter.ClientSecret,
	}))

}

// SetupSEO add seo
func SetupSEO(Admin *admin.Admin) {
	models.SEOCollection = qor_seo.New("Common SEO")
	models.SEOCollection.RegisterGlobalVaribles(&models.SEOGlobalSetting{SiteName: "QorPress"})
	models.SEOCollection.SettingResource = Admin.AddResource(&models.SEOSetting{}, &admin.Config{Invisible: true})
	models.SEOCollection.RegisterSEO(&qor_seo.SEO{
		Name: "Default Page",
	})
	models.SEOCollection.RegisterSEO(&qor_seo.SEO{
		Name:     "Post",
		Varibles: []string{"Name", "CategoryName"},
		Context: func(objects ...interface{}) map[string]string {
			post := objects[0].(models.Post)
			context := make(map[string]string)
			context["Title"] = post.Title
			context["CategoryName"] = post.Categories[0].Name
			return context
		},
	})
	models.SEOCollection.RegisterSEO(&qor_seo.SEO{
		Name:     "Category",
		Varibles: []string{"Name"},
		Context: func(objects ...interface{}) map[string]string {
			category := objects[0].(models.Category)
			context := make(map[string]string)
			context["Name"] = category.Name
			return context
		},
	})
	models.SEOCollection.RegisterSEO(&qor_seo.SEO{
		Name:     "Tag",
		Varibles: []string{"Name"},
		Context: func(objects ...interface{}) map[string]string {
			tag := objects[0].(models.Tag)
			context := make(map[string]string)
			context["Name"] = tag.Name
			return context
		},
	})
	Admin.AddResource(models.SEOCollection, &admin.Config{
		Name:      "SEO Setting",
		Menu:      []string{"Site Management"},
		Singleton: true,
		Priority:  2,
	},
	)
}

func SetupDB() {
	db = services.Init()
	db.LogMode(true)
	db.AutoMigrate(
		// generic
		&models.Article{},
		&models.Page{},
		&models.User{},
		&models.Comment{},
		&models.Category{},
		&models.Post{},
		&models.Tag{},
		&models.Page{},
		&models.Document{},
		&models.Video{},
		&models.Image{},
		&models.Link{},
		&models.AuthIdentity{},
		&models.SignLog{},
		&models.AuthInfo{},
		&models.SEOSetting{},
		&asset_manager.AssetManager{},
		&i18n_database.Translation{},
		&models.WidgetSetting{},
		// ex custom plugin for oniontree, need to loaded through a go plugin. need to be investigated
		// &models.Service{},
		// &models.URL{},
		// &models.PublicKey{},
	)

	var post models.Post
	var video []models.Video
	var image []models.Image
	var link []models.Link
	var documents []models.Document

	db.Model(&post).Related(&video)
	db.Model(&post).Related(&image)
	db.Model(&post).Related(&link)
	db.Model(&post).Related(&documents)

	I18n = i18n.New(
		database.New(db), // load translations from the database
		yaml.New(filepath.Join(config.Root, "shared/locales")), // load translations from the YAML files in directory `config/locales`
	)

	media.RegisterCallbacks(db)
	l10n.RegisterCallbacks(db)
	sorting.RegisterCallbacks(db)
	validations.RegisterCallbacks(db)
	media.RegisterCallbacks(db)
	publish2.RegisterCallbacks(db)
}

func setupTemplateFuncs() template.FuncMap {
	funcMaps := sprig.FuncMap()
	funcMaps["unsafeHtml"] = utils.UnsafeHtml
	funcMaps["stripSummaryTags"] = utils.StripSummaryTags
	funcMaps["displayDateString"] = utils.DisplayDateString
	funcMaps["displayDate"] = utils.DisplayDateV2
	funcMaps["truncateBody"] = utils.TruncateBody

	gtf.Inject(funcMaps)
	return funcMaps
}

func RequestIdMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		xuid, _ := uuid.NewV4()
		c.Writer.Header().Set("X-Request-Id", xuid.String())
		c.Next()
	}
}

func SetupAdmin() {
}

func SetupRouter() *gin.Engine {
	mux := http.NewServeMux()

	mux.Handle("/auth/", Auth.NewServeMux())

	Admin := admin.New(&admin.AdminConfig{DB: db})

	// Initialize AssetFS
	AssetFS := assetfs.AssetFS().NameSpace("admin")
	// AssetFS := Admin.SetAssetFS(bindatafs.AssetFS.NameSpace("admin"))
	// Register custom paths to manually saved views
	AssetFS.RegisterPath(filepath.Join(qor_utils.AppRoot, "qor/admin/views"))

	Admin.MountTo("/admin", mux)

	//API Setup
	API := admin.New(&qor.Config{
		DB: db.Set(publish2.VisibleMode, publish2.ModeOff).Set(publish2.ScheduleMode, publish2.ModeOff),
	})

	API.AddResource(&models.Post{})
	API.AddResource(&models.Comment{})
	API.AddResource(&models.Image{})
	API.AddResource(&models.Video{})
	API.AddResource(&models.Category{})
	API.AddResource(&models.Link{})
	API.AddResource(&models.Tag{})

	API.MountTo("/adminapi", mux)

	assetManager := Admin.AddResource(&asset_manager.AssetManager{}, &admin.Config{Invisible: true})

	Admin.AddResource(&media_library.MediaLibrary{}, &admin.Config{
		Menu: []string{"Site Management"},
	})

	Admin.AddResource(&models.User{}, &admin.Config{
		Name: "Users",
		Menu: []string{"Site Management"},
	})

	// Add Translations
	Admin.AddResource(I18n, &admin.Config{
		Menu:     []string{"Site Management"},
		Priority: -1,
	})

	category := Admin.AddResource(&models.Category{}, &admin.Config{
		Name: "Categories",
		Menu: []string{"Content Management"},
	})
	category.Meta(&admin.Meta{
		Name: "Categories",
		Type: "select_many",
	})

	// Add ProductImage as Media Libraray
	postImagesResource := Admin.AddResource(&models.Post{}, &admin.Config{Menu: []string{"Post Management"}, Priority: -1})

	postImagesResource.Filter(&admin.Filter{
		Name:       "SelectedType",
		Label:      "Media Type",
		Operations: []string{"contains"},
		Config:     &admin.SelectOneConfig{Collection: [][]string{{"video", "Video"}, {"image", "Image"}, {"file", "File"}, {"video_link", "Video Link"}}},
	})

	postImagesResource.Filter(&admin.Filter{
		Name:   "Category",
		Config: &admin.SelectOneConfig{RemoteDataResource: category},
	})
	postImagesResource.IndexAttrs("File", "Title")

	post := Admin.AddResource(&models.Post{}, &admin.Config{
		Name: "Posts",
		Menu: []string{"Content Management"},
	})

	// post.Meta(&admin.Meta{Name: "Category", Config: &admin.SelectOneConfig{Collection: Category, AllowBlank: true}})

	/*
	post.Meta(&admin.Meta{Name: "Body", Config: &admin.RichEditorConfig{Plugins: []admin.RedactorPlugin{
		{Name: "medialibrary", Source: "/admin/assets/javascripts/qor_redactor_medialibrary.js"},
		{Name: "table", Source: "/vendors/redactor_table.js"},
	},
		Settings: map[string]interface{}{
			"medialibraryUrl": "/admin/product_images",
		},
	}})
	*/
	// post.Meta(&admin.Meta{Name: "Category", Config: &admin.SelectOneConfig{AllowBlank: true}})
	post.Meta(&admin.Meta{Name: "Categories", Config: &admin.SelectManyConfig{SelectMode: "bottom_sheet"}})

	post.Meta(&admin.Meta{Name: "MainImage", Config: &media_library.MediaBoxConfig{
		RemoteDataResource: postImagesResource,
		Max:                1,
		Sizes: map[string]*media.Size{
			"main": {Width: 560, Height: 700},
		},
	}})

	post.Meta(&admin.Meta{Name: "MainImageURL", Valuer: func(record interface{}, context *qor.Context) interface{} {
		if p, ok := record.(*models.Post); ok {
			result := bytes.NewBufferString("")
			tmpl, _ := template.New("").Parse("<hr><img src='{{.image}}'></img><hr>")
			tmpl.Execute(result, map[string]string{"image": p.MainImageURL()})
			return template.HTML(result.String())
		}
		return ""
	}})

	postPropertiesRes := post.Meta(&admin.Meta{Name: "PostProperties"}).Resource
	postPropertiesRes.NewAttrs(&admin.Section{
		Rows: [][]string{{"Name", "Value"}},
	})
	postPropertiesRes.EditAttrs(&admin.Section{
		Rows: [][]string{{"Name", "Value"}},
	})

	post.Filter(&admin.Filter{
		Name:   "Category",
		Config: &admin.SelectOneConfig{RemoteDataResource: category},
	})

	post.Filter(&admin.Filter{
		Name: "Title",
		Type: "string",
	})

	post.Filter(&admin.Filter{
		Name: "Slug",
	})

	post.Filter(&admin.Filter{
		Name: "CreatedAt",
	})

	post.Action(&admin.Action{
		Name:        "Import Post",
		URLOpenType: "slideout",
		URL: func(record interface{}, context *admin.Context) string {
			return "/admin/workers/new?job=Import Products"
		},
		Modes: []string{"collection"},
	})

	post.IndexAttrs("ID", "Title", "Summary", "MainImage", "Type", "Categories")
	post.NewAttrs("Title", "Summary", "Images", "Videos", "Links", "Documents", "Categories", "Type")
	post.Meta(&admin.Meta{
		Name: "Body",
		Config: &admin.RichEditorConfig{
			AssetManager: assetManager,
		},
	})

	post.Meta(&admin.Meta{
		Name: "Type",
		Type: "select_one",
		Config: &admin.SelectOneConfig{
			Collection: []string{"article", "publication", "blog", "video", "press_release", "event", "news"},
		},
	})

	Admin.AddResource(&models.Tag{}, &admin.Config{
		Name: "Tags",
		Menu: []string{"Content Management"},
	})

	Admin.AddResource(&models.Link{}, &admin.Config{
		Name: "Links",
		Menu: []string{"Content Management"},
	})

	Admin.AddResource(&models.Image{}, &admin.Config{
		Name: "Images",
		Menu: []string{"Content Management"},
	})

	Admin.AddResource(&models.Document{}, &admin.Config{
		Name: "Documents",
		Menu: []string{"Content Management"},
	})

	Admin.AddResource(&models.Video{}, &admin.Config{
		Name: "Videos",
		Menu: []string{"Content Management"},
	})

	// Setup pages
	PageBuilderWidgets := widget.New(&widget.Config{DB: db})
	PageBuilderWidgets.WidgetSettingResource = Admin.NewResource(&models.WidgetSetting{}, &admin.Config{Name: "PageBuilderWidgets"})
	PageBuilderWidgets.WidgetSettingResource.NewAttrs(
		&admin.Section{
			Rows: [][]string{{"Kind"}, {"SerializableMeta"}},
		},
	)
	PageBuilderWidgets.WidgetSettingResource.AddProcessor(&resource.Processor{
		Handler: func(value interface{}, metaValues *resource.MetaValues, context *qor.Context) error {
			if widgetSetting, ok := value.(*models.WidgetSetting); ok {
				if widgetSetting.Name == "" {
					var count int
					context.GetDB().Set(admin.DisableCompositePrimaryKeyMode, "off").Model(&models.WidgetSetting{}).Count(&count)
					widgetSetting.Name = fmt.Sprintf("%v %v", qor_utils.ToString(metaValues.Get("Kind").Value), count)
				}
			}
			return nil
		},
	})
	Admin.AddResource(PageBuilderWidgets, &admin.Config{Menu: []string{"Pages Management"}})

	page := page_builder.New(&page_builder.Config{
		Admin:       Admin,
		PageModel:   &models.Page{},
		Containers:  PageBuilderWidgets,
		AdminConfig: &admin.Config{Name: "Pages", Menu: []string{"Pages Management"}, Priority: 1},
	})
	page.IndexAttrs("ID", "Title", "PublishLiveNow")

	PostExchange := exchange.NewResource(&models.Post{}, exchange.Config{PrimaryField: "Title"})
	// PostExchange.Meta(&exchange.Meta{Name: "Slug"})
	PostExchange.Meta(&exchange.Meta{Name: "Title"})
	PostExchange.Meta(&exchange.Meta{Name: "Body"})

	PostExchange.AddValidator(&resource.Validator{
		Handler: func(record interface{}, metaValues *resource.MetaValues, context *qor.Context) error {
			//if utils.ToInt(metaValues.Get("Price").Value) < 100 {
			//	return validations.NewError(record, "Price", "price can't less than 100")
			//}
			return nil
		},
	})

	Worker := worker.New()

	type sendNewsletterArgument struct {
		Subject      string
		Content      string `sql:"size:65532"`
		SendPassword string
		worker.Schedule
	}

	Worker.RegisterJob(&worker.Job{
		Name: "Send Newsletter",
		Handler: func(argument interface{}, qorJob worker.QorJobInterface) error {
			qorJob.AddLog("Started sending newsletters...")
			qorJob.AddLog(fmt.Sprintf("Argument: %+v", argument.(*sendNewsletterArgument)))
			for i := 1; i <= 100; i++ {
				time.Sleep(100 * time.Millisecond)
				qorJob.AddLog(fmt.Sprintf("Sending newsletter %v...", i))
				qorJob.SetProgress(uint(i))
			}
			qorJob.AddLog("Finished send newsletters")
			return nil
		},
		Resource: Admin.NewResource(&sendNewsletterArgument{}),
	})

	type importProductArgument struct {
		File oss.OSS
	}

	Worker.RegisterJob(&worker.Job{
		Name:  "Import Posts",
		Group: "Posts Management",
		Handler: func(arg interface{}, qorJob worker.QorJobInterface) error {
			argument := arg.(*importProductArgument)

			context := &qor.Context{DB: db}

			var errorCount uint

			if err := PostExchange.Import(
				csv.New(filepath.Join("public", argument.File.URL())),
				context,
				func(progress exchange.Progress) error {
					var cells = []worker.TableCell{
						{Value: fmt.Sprint(progress.Current)},
					}

					var hasError bool
					for _, cell := range progress.Cells {
						var tableCell = worker.TableCell{
							Value: fmt.Sprint(cell.Value),
						}

						if cell.Error != nil {
							hasError = true
							errorCount++
							tableCell.Error = cell.Error.Error()
						}

						cells = append(cells, tableCell)
					}

					if hasError {
						if errorCount == 1 {
							var headerCells = []worker.TableCell{
								{Value: "Line No."},
							}
							for _, cell := range progress.Cells {
								headerCells = append(headerCells, worker.TableCell{
									Value: cell.Header,
								})
							}
							qorJob.AddResultsRow(headerCells...)
						}

						qorJob.AddResultsRow(cells...)
					}

					qorJob.SetProgress(uint(float32(progress.Current) / float32(progress.Total) * 100))
					qorJob.AddLog(fmt.Sprintf("%d/%d Importing post %v", progress.Current, progress.Total, progress.Value.(*models.Post).Title))
					return nil
				},
			); err != nil {
				qorJob.AddLog(err.Error())
			}

			return nil
		},
		Resource: Admin.NewResource(&importProductArgument{}),
	})

	Worker.RegisterJob(&worker.Job{
		Name:  "Export Posts",
		Group: "Posts Management",
		Handler: func(arg interface{}, qorJob worker.QorJobInterface) error {
			qorJob.AddLog("Exporting products...")

			context := &qor.Context{DB: db}
			fileName := fmt.Sprintf("/downloads/products.%v.csv", time.Now().UnixNano())
			if err := PostExchange.Export(
				csv.New(filepath.Join("public", fileName)),
				context,
				func(progress exchange.Progress) error {
					qorJob.AddLog(fmt.Sprintf("%v/%v Exporting product %v", progress.Current, progress.Total, progress.Value.(*models.Post).Title))
					return nil
				},
			); err != nil {
				qorJob.AddLog(err.Error())
			}

			qorJob.SetProgressText(fmt.Sprintf("<a href='%v'>Download exported products</a>", fileName))
			return nil
		},
	})

	exchange_actions.RegisterExchangeJobs(I18n, Worker)
	Admin.AddResource(Worker, &admin.Config{Menu: []string{"Site Management"}, Priority: 3})

	SetupSEO(Admin)

	router := gin.Default()

	if runtime.GOOS == "linux" {
		log.Println("Loading html from binary")
		router.SetHTMLTemplate(templates)
	}

	if runtime.GOOS == "darwin" {
		router.SetFuncMap(setupTemplateFuncs())
		router.LoadHTMLGlob("views/**/*")
	}

	// Ping handler
	router.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	router.GET("/stats", func(context *gin.Context) {
		context.JSON(http.StatusOK, stats.Report())
	})
	router.GET("/", controllers.Home)
	router.GET("/about-us", controllers.AboutUs)
	router.GET("/categories/:category", controllers.GetPostsForCategory)
	router.GET("blog/:year/:month/:day/:slug", controllers.GetPost)
	router.GET("posts/:slug", controllers.GetPost)
	router.GET("/news", controllers.GetNews)
	router.GET("/news/:page", controllers.GetNews)

	router.GET("/tags", func(c *gin.Context) {
	    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "3"))
	    var tags []models.Tag
	    paginator := pagination.Paging(&pagination.Param{
	        DB:      db,
	        Page:    page,
	        Limit:   limit,
	        OrderBy: []string{"id desc"},
	        ShowSQL: true,
	    }, &tags)
	    c.IndentedJSON(200, paginator)
	})

	// router.Static("/public", "./public")
	// router.Static("/assets", "./assets")

	url := ginSwagger.URL("http://localhost:4000/swagger/doc.json") // The url pointing to API definition
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	//API Calls
	//api := router.Group("/api")
	//{
	// api.GET("/get-tweets", controllers.GetTweets)
	// api.GET("/get-flickr", controllers.GetFlickr)
	// api.GET("/testdata", controllers.GetTestData)
	//}

	router.Static("/system", "./public/system")
	router.Static("/public", "./public")

	admin := router.Group("/admin", gin.BasicAuth(
		gin.Accounts{
			"admin":   "admin",
			"x0rzkov": "x0rzkov",
		}))
	{
		admin.Any("/*resources", gin.WrapH(mux))
	}
	router.Any("/adminapi/*resources", gin.WrapH(mux))
	router.NoRoute(func(context *gin.Context) {
		fmt.Println(">>>>>>>>>>>>>>>>>> 404 <<<<<<<<<<<<<<<<<<<")
		context.HTML(http.StatusNotFound, "content_not_found", nil)
	})
	return router
}

func loadTemplates() (*template.Template, error) {
	templates = template.New("")
	templates.Funcs(setupTemplateFuncs())
	var myAssets = Assets.Files

	for name, file := range myAssets {
		if file.IsDir() || !strings.HasSuffix(name, ".html") {
			continue
		}
		h, err := ioutil.ReadAll(file)
		if err != nil {
			return nil, err
		}
		templates, err = templates.New(name).Parse(string(h))
		if err != nil {
			return nil, err
		}
	}

	return templates, nil
}

func main() {
	port := flag.String("port", "4000", "The port the app will listen to")
	host := flag.String("host", "0.0.0.0", "The ip address to listen on")
	compileTemplate := flag.Bool("compile-templates", false, "Set this to true to compile templates to binary")

	flag.Parse()

	if *compileTemplate {
		Admin := admin.New(&admin.AdminConfig{
			DB:      db.Set(publish2.VisibleMode, publish2.ModeOff).Set(publish2.ScheduleMode, publish2.ModeOff),
			AssetFS: bindatafs.AssetFS.NameSpace("admin")})
		Admin.SetAssetFS(bindatafs.AssetFS.NameSpace("admin"))
		bindatafs.AssetFS.Compile()
	} else {
		SetupDB()
		defer db.Close()

		loadTemplates()
		r := SetupRouter()
		fmt.Println(*host, *port)

		m := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist("x0rzkov.com", "localhost.com"),
			Cache:      autocert.DirCache("./shared/cache"),
		}

		// http.ListenAndServe(fmt.Sprintf("%s:%d", "", config.Config.App.Port), manager.SessionManager.Middleware(RedirectBack.Middleware(mux)))

		if runtime.GOOS == "linux" {
			log.Fatal(autotls.RunWithManager(r, &m))
		} else {
			r.Run(fmt.Sprintf("%s:%s", *host, *port))
		}
	}
}
