package controllers

import (
	"net/http"

	"github.com/qorpress/qorpress-contrib/twitter/models"
	"github.com/qorpress/qorpress/core/render"
	"github.com/qorpress/qorpress/pkg/utils"
)

// Controller posts controller
type Controller struct {
	View *render.Render
}

// Index posts index page
func (ctrl Controller) Index(w http.ResponseWriter, req *http.Request) {
	var (
		Tweets []models.TwitterTweet
		tx     = utils.GetDB(req)
	)
	tx.Find(&Tweets)
	ctrl.View.Execute("index", map[string]interface{}{}, req, w)
}
