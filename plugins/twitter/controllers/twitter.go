package controllers

import (
	"fmt"
	"path/filepath"

	"github.com/qorpress/qorpress/core/admin"

	// "github.com/dghubble/oauth1"
	// "github.com/koreset/go-twitter/twitter"
	// "github.com/qorpress/qorpress/core/media"
	// "github.com/qorpress/qorpress/core/media/media_library"
	// "github.com/qorpress/qorpress/core/qor"
	"github.com/qorpress/qorpress-contrib/twitter/models"
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
	themeDir := fmt.Sprintf(filepath.Join(config.Root, "plugins", "twitter", "views"))
	controller := &Controller{View: render.New(&render.Config{
		AssetFileSystem: application.AssetFS.NameSpace("twitter"),
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
	Admin.AddResource(&models.TwitterSetting{}, &admin.Config{Name: "API Parameters", Menu: []string{"Twitter Management"}, Singleton: true, Priority: 1})
}

/*
func GetTweets(c *gin.Context) {
	cfg := oauth1.NewConfig(config.Config.ApiKey.Twitter.ConsumerKey, config.Config.ApiKey.Twitter.ConsumerSecret)
	token := oauth1.NewToken(config.Config.ApiKey.Twitter.AccessToken, config.Config.ApiKey.Twitter.AccessSecret)
	httpClient := cfg.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	tweets, response, err := client.Timelines.UserTimeline(&twitter.UserTimelineParams{
		ScreenName: "lucmichalski",
		Count:      5,
		TweetMode:  "extended",
	})
	shallowTweets := utils.GetShallowTweets(tweets)
	if err != nil {
		panic(err.Error())
	}
	if response.StatusCode == http.StatusOK {
		c.JSON(http.StatusOK, shallowTweets)
	} else {
		c.JSON(http.StatusOK, "test")
	}
}
*/

