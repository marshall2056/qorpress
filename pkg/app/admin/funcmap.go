package admin

import (
	"html/template"

	"github.com/qorpress/qorpress/internal/admin"
)

func initFuncMap(Admin *admin.Admin) {
	Admin.RegisterFuncMap("render_latest_posts", renderLatestPost)
}

func renderLatestPost(context *admin.Context) template.HTML {
	var postContext = context.NewResourceContext("Post")
	postContext.Searcher.Pagination.PerPage = 5
	// postContext.SetDB(postContext.GetDB().Where("state in (?)", []string{"paid"}))

	if posts, err := postContext.FindMany(); err == nil {
		return postContext.Render("index/table", posts)
	}
	return template.HTML("")
}
