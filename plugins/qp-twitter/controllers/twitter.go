package controllers

import (
	"net/http"

	"github.com/dghubble/oauth1"
	"github.com/gin-gonic/gin"
	"github.com/koreset/go-twitter/twitter"

	"github.com/qorpress/qorpress/pkg/config"
	"github.com/qorpress/qorpress/pkg/utils"
)

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
