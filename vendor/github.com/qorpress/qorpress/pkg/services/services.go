package services

import (
	"fmt"

	"github.com/k0kubun/pp"
	"github.com/kennygrant/sanitize"

	"github.com/qorpress/qorpress/pkg/models"
)

var defaultTags = []string{"h1", "h2", "h3", "h4", "h5", "h6", "div", "span", "hr", "p", "br", "b", "i", "strong", "em", "ol", "ul", "li", "a", "img", "pre", "code", "blockquote", "article", "section"}

var defaultAttributes = []string{"id", "src", "href", "title", "alt", "name", "rel"}

func GetPosts(start, limit int) []models.Post {
	var posts []models.Post
	GetDB().Set("gorm:auto_preload", true).Where("type in (?) and body != ''", []string{"article", "press_release", "news"}).Order("created desc").Offset(start).Limit(limit).Find(&posts)
	// Lets sanitize the html output and strip off MSOffice tags
	for _, post := range posts {
		post.Body, _ = sanitize.HTMLAllowing(post.Body, defaultTags, defaultAttributes)
		//fmt.Println(post.Images[0].File.OSS.URL("large"))
	}
	return posts
}

func GetEvents(start, limit int) []models.Post {
	var posts []models.Post
	GetDB().Set("gorm:auto_preload", true).Where("type in (?) and body != ''", []string{"event"}).Order("created desc").Offset(start).Limit(limit).Find(&posts)
	// Lets sanitize the html output and strip off MSOffice tags
	for _, post := range posts {
		post.Body, _ = sanitize.HTMLAllowing(post.Body, defaultTags, defaultAttributes)
		//fmt.Println(post.Images[0].File.OSS.URL("large"))
	}
	return posts
}

func GetRecentPosts(start, limit int) []models.Post {
	var posts []models.Post
	GetDB().Where("type in (?) and body != '' and summary != ''", []string{"article", "press_release", "news"}).Preload("Images").Preload("Categories").Order("created desc").Offset(start).Limit(limit).Find(&posts)
	// Lets sanitize the html output and strip off MSOffice tags
	for _, post := range posts {
		post.Body, _ = sanitize.HTMLAllowing(post.Body, defaultTags, defaultAttributes)
		//fmt.Println(post.Images[0].File.OSS.URL("large"))
	}
	return posts
}

func GetVideos() []models.Post {
	var videos []models.Post
	GetDB().Where("type = 'video'").Preload("Images").Preload("Videos").Order("created desc").Find(&videos)
	return videos
}

func GetPost(postid int) models.Post {
	var post models.Post
	GetDB().Set("gorm:auto_preload", true).Where("id = ? ", postid).First(&post)

	pp.Println(post)
	return post
}

func GetPostBySlug(slug string) models.Post {
	var post models.Post
	GetDB().Set("gorm:auto_preload", true).Where("uuid = ? ", slug).First(&post)
	return post
}

func GetPublications(start, limit int) []models.Post {
	var publications []models.Post

	GetDB().Where("type = 'publication'").Preload("Images").Preload("Links").Preload("Documents").Order("created desc").Offset(start).Limit(limit).Find(&publications)

	return publications
}

func GetPublication(postid int) models.Post {
	var pub models.Post
	GetDB().Where("id = ?", postid).Preload("Images").Preload("Videos").Preload("Links").First(&pub)
	return pub
}

func GetPostsForCategory(start int, limit int, categoryString string) []models.Post {
	var category models.Category
	var posts []models.Post
	GetDB().First(&category, "name = ?", categoryString)
	GetDB().Set("gorm:auto_preload", true).Model(&category).Order("created desc").Offset(start).Limit(limit).Related(&posts, "Posts")
	fmt.Println(posts)
	return posts
}
