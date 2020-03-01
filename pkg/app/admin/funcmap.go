package admin

import (
	"html/template"

	"github.com/qorpress/admin"
)

func initFuncMap(Admin *admin.Admin) {
	Admin.RegisterFuncMap("render_latest_order", renderLatestOrder)
	Admin.RegisterFuncMap("render_latest_posts", renderLatestPost)
}

func renderLatestOrder(context *admin.Context) template.HTML {
	var orderContext = context.NewResourceContext("Order")
	orderContext.Searcher.Pagination.PerPage = 5
	// orderContext.SetDB(orderContext.GetDB().Where("state in (?)", []string{"paid"}))

	if orders, err := orderContext.FindMany(); err == nil {
		return orderContext.Render("index/table", orders)
	}
	return template.HTML("")
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
