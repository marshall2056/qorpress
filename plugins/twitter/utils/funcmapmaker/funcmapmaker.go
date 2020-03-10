package funcmapmaker

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/dghubble/oauth1"
	"github.com/koreset/go-twitter/twitter"

	"github.com/qorpress/qorpress-contrib/twitter/models"
	"github.com/qorpress/qorpress/core/render"
	"github.com/qorpress/qorpress/pkg/utils"
)

func AddFuncMapMaker(view *render.Render) *render.Render {
	oldFuncMapMaker := view.FuncMapMaker
	view.FuncMapMaker = func(render *render.Render, req *http.Request, w http.ResponseWriter) template.FuncMap {
		funcMap := template.FuncMap{}
		if oldFuncMapMaker != nil {
			funcMap = oldFuncMapMaker(render, req, w)
		}

		fmt.Println("adding funcMap[\"get_twitter_screename\"]")
		funcMap["get_twitter_screename"] = func() string {
			var ts models.TwitterSetting
			utils.GetDB(req).Find(&ts)
			return ts.ScreenName
		}

		fmt.Println("adding funcMap[\"get_tweets\"]")
		funcMap["get_tweets"] = func() (shallowTweets []models.TwitterShallowTweet) {
			var ts models.TwitterSetting
			utils.GetDB(req).Find(&ts)
			if ts.ConsumerKey != "" && ts.ConsumerSecret != "" && ts.AccessToken != "" && ts.AccessSecret != "" {
				cfg := oauth1.NewConfig(ts.ConsumerKey, ts.ConsumerSecret)
				token := oauth1.NewToken(ts.AccessToken, ts.AccessSecret)
				httpClient := cfg.Client(oauth1.NoContext, token)
				client := twitter.NewClient(httpClient)
				tweets, _, err := client.Timelines.UserTimeline(&twitter.UserTimelineParams{
					ScreenName: ts.ScreenName,
					Count:      ts.Count,
					TweetMode:  "extended",
				})
				shallowTweets = models.GetShallowTweets(tweets)
				if err != nil {
					panic(err.Error())
				}
			}
			return
		}
		fmt.Println("adding funcMap[\"get_twitter_profile\"]")
		funcMap["get_twitter_profile"] = func() (profile models.TwitterProfile) {
			query := fmt.Sprintf(`SELECT T.* FROM (POST_TAGS ST, TAGS T) WHERE ST.tag_id=t.id`)
			utils.GetDB(req).Raw(query).Scan(&profile)
			fmt.Println(query)
			return
		}

		return funcMap
	}

	return view
}
