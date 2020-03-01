package posts

import (
	"net/http"
	"strings"

	"github.com/qorpress/render"

	"github.com/qorpress/qorpress-example/pkg/models/posts"
	"github.com/qorpress/qorpress-example/pkg/utils"
)

// Controller posts controller
type Controller struct {
	View *render.Render
}

// Index posts index page
func (ctrl Controller) Index(w http.ResponseWriter, req *http.Request) {
	var (
		Posts []posts.Post
		tx       = utils.GetDB(req)
	)

	tx.Preload("Category").Find(&Posts)

	ctrl.View.Execute("index", map[string]interface{}{}, req, w)
}

// Show post show page
func (ctrl Controller) Show(w http.ResponseWriter, req *http.Request) {
	var (
		post        posts.Post
		// colorVariation posts.ColorVariation
		codes          = strings.Split(utils.URLParam("code", req), "_")
		postCode    = codes[0]
		// colorCode      string
		tx             = utils.GetDB(req)
	)

	//if len(codes) > 1 {
	//	colorCode = codes[1]
	//}

	if tx.Preload("Category").Where(&posts.Post{Code: postCode}).First(&post).RecordNotFound() {
		http.Redirect(w, req, "/", http.StatusFound)
	}

	// tx.Preload("Post").Preload("Color").Preload("SizeVariations.Size").Where(&posts.ColorVariation{PostID: post.ID, ColorCode: colorCode}).First(&colorVariation)
	// ctrl.View.Execute("show", map[string]interface{}{"CurrentColorVariation": colorVariation}, req, w)
}

// Category category show page
func (ctrl Controller) Category(w http.ResponseWriter, req *http.Request) {
	var (
		category posts.Category
		Posts []posts.Post
		tx       = utils.GetDB(req)
	)

	if tx.Where("code = ?", utils.URLParam("code", req)).First(&category).RecordNotFound() {
		http.Redirect(w, req, "/", http.StatusFound)
	}

	// tx.Where(&posts.Post{CategoryID: category.ID}).Preload("ColorVariations").Find(&Posts)

	ctrl.View.Execute("category", map[string]interface{}{"CategoryName": category.Name, "Posts": Posts}, req, w)
}
