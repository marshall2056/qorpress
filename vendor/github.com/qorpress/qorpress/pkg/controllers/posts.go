package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/qorpress/qorpress/pkg/services"
)

func GetPost(c *gin.Context) {
	payload := make(map[string]interface{})
	slug := c.Param("slug")
	fmt.Println("The Slug: ", slug)
	if slug == "" {
		c.HTML(http.StatusNotFound, "content_not_found", nil)
		return
	}
	post := services.GetPostBySlug(slug)
	recentPosts := services.GetRecentPosts(0, 5)
	payload["post"] = post
	payload["recentPosts"] = recentPosts
	payload["active"] = "none"
	payload["title"] = post.Title
	c.HTML(http.StatusOK, "post-detail", payload)
}

func GetPublications(c *gin.Context) {
	c.HTML(http.StatusOK, "publications_home", nil)
}

func GetPostsForCategory(c *gin.Context) {
	payload := make(map[string]interface{})
	categoryTitle := c.Param("category")
	payload["categoryTitle"] = categoryTitle
	c.HTML(http.StatusOK, "category-list", payload)
}
