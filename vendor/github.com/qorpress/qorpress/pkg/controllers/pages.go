package controllers

/*
import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/qorpress/qorpress/pkg/services"
	"github.com/qorpress/qorpress/pkg/utils"
)

func GetPage(c *gin.Context) {
	payload := make(map[string]interface{})
	slug := c.Param("slug")
	fmt.Println("The Slug: ", slug)
	if slug == "" {
		c.HTML(http.StatusNotFound, "content_not_found", nil)
		return
	}
	post := services.GetPageBySlug(slug)
	payload["post"] = post
	payload["active"] = "none"
	payload["title"] = post.Title
	c.HTML(http.StatusOK, "page-detail", payload)
}
*/