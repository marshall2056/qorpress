package controllers

/*
import (
	"github.com/biezhi/gorm-paginator/pagination"
)

func GetTag(c *gin.Context) {
	payload := make(map[string]interface{})
	slug := c.Param("slug")
	fmt.Println("The Slug: ", slug)
	if slug == "" {
		c.HTML(http.StatusNotFound, "content_not_found", nil)
		return
	}
	post := services.GetTagBySlug(slug)
	payload["post"] = post
	payload["active"] = "none"
	payload["title"] = post.Title
	c.HTML(http.StatusOK, "page-detail", payload)
}
*/