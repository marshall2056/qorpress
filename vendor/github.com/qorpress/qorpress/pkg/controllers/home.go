package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/qorpress/qorpress/pkg/services"
)

func Home(c *gin.Context) {
	payload := make(map[string]interface{})
	postsTop := services.GetPosts(0, 3)
	postsBottom := services.GetPosts(4, 3)
	videos := services.GetVideos()
	payload["postsTop"] = postsTop
	payload["postsBottom"] = postsBottom
	payload["videos"] = videos
	payload["active"] = "home_page"
	payload["title"] = "Home"
	c.HTML(http.StatusOK, "home", payload)
}
