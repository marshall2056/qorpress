package funcmapmaker

import (
	"fmt"
	"html/template"
	"net/http"

	// "github.com/k0kubun/pp"
	"github.com/qorpress/qorpress/core/action_bar"
	"github.com/qorpress/qorpress/core/i18n/inline_edit"
	"github.com/qorpress/qorpress/core/qor"
	"github.com/qorpress/qorpress/core/render"
	"github.com/qorpress/qorpress/core/session"
	"github.com/qorpress/qorpress/core/session/manager"
	"github.com/qorpress/qorpress/core/widget"
	"github.com/qorpress/qorpress/pkg/app/admin"
	"github.com/qorpress/qorpress/pkg/config/i18n"
	"github.com/qorpress/qorpress/pkg/models/posts"
	"github.com/qorpress/qorpress/pkg/models/seo"
	"github.com/qorpress/qorpress/pkg/models/users"
	"github.com/qorpress/qorpress/pkg/utils"
)

// GetEditMode get edit mode
func GetEditMode(w http.ResponseWriter, req *http.Request) bool {
	return admin.ActionBar.EditMode(w, req)
}

// AddFuncMapMaker add FuncMapMaker to view
func AddFuncMapMaker(view *render.Render) *render.Render {
	oldFuncMapMaker := view.FuncMapMaker
	view.FuncMapMaker = func(render *render.Render, req *http.Request, w http.ResponseWriter) template.FuncMap {
		funcMap := template.FuncMap{}
		if oldFuncMapMaker != nil {
			funcMap = oldFuncMapMaker(render, req, w)
		}

		// Add `t` method
		for key, fc := range inline_edit.FuncMap(i18n.I18n, utils.GetCurrentLocale(req), GetEditMode(w, req)) {
			funcMap[key] = fc
		}

		for key, value := range admin.ActionBar.FuncMap(w, req) {
			funcMap[key] = value
		}

		widgetContext := admin.Widgets.NewContext(&widget.Context{
			DB:         utils.GetDB(req),
			Options:    map[string]interface{}{"Request": req},
			InlineEdit: GetEditMode(w, req),
		})
		for key, fc := range widgetContext.FuncMap() {
			funcMap[key] = fc
		}

		fmt.Println("adding funcMap[\"raw\"]")
		funcMap["raw"] = func(str string) template.HTML {
			return template.HTML(utils.HTMLSanitizer.Sanitize(str))
		}

		fmt.Println("adding funcMap[\"flashes\"]")
		funcMap["flashes"] = func() []session.Message {
			return manager.SessionManager.Flashes(w, req)
		}

		// Add `action_bar` method
		fmt.Println("adding funcMap[\"render_action_bar\"]")
		funcMap["render_action_bar"] = func() template.HTML {
			return admin.ActionBar.Actions(action_bar.Action{
				Name: "Edit SEO",
				Link: seo.SEOCollection.SEOSettingURL("/help"),
			}).Render(w, req)
		}

		fmt.Println("adding funcMap[\"render_seo_tag\"]")
		funcMap["render_seo_tag"] = func() template.HTML {
			return seo.SEOCollection.Render(&qor.Context{DB: utils.GetDB(req)}, "Default Page")
		}

		fmt.Println("adding funcMap[\"get_categories\"]")
		funcMap["get_categories"] = func() (categories []posts.Category) {
			utils.GetDB(req).Find(&categories)
			return
		}

		fmt.Println("adding funcMap[\"get_post_tags\"]")
		funcMap["get_post_tags"] = func(postId uint) (tags []posts.Tag) {
			query := fmt.Sprintf(`SELECT T.* FROM (POST_TAGS ST, TAGS T) WHERE ST.POST_ID=%d AND ST.tag_id=t.id`, postId)
			utils.GetDB(req).Raw(query).Scan(&tags)
			// pp.Println("tags:", tags)
			fmt.Println(query)
			return
		}

		fmt.Println("adding funcMap[\"get_category_tags\"]")
		funcMap["get_category_tags"] = func(catId uint) (tags []posts.Tag) {
			query := fmt.Sprintf(`
				SELECT T.*
				FROM (POSTS S, TAGS T)
				INNER JOIN POST_TAGS ST ON S.ID = ST.POST_ID 
				INNER JOIN TAGS ON ST.TAG_ID = T.ID 
				WHERE S.CATEGORY_ID=%d LIMIT 100`, catId)
			utils.GetDB(req).Raw(query).Scan(&tags)
			// pp.Println("tags:", tags)
			// fmt.Println(query)
			return
		}

		fmt.Println("adding funcMap[\"current_locale\"]")
		funcMap["current_locale"] = func() string {
			return utils.GetCurrentLocale(req)
		}

		fmt.Println("adding funcMap[\"current_user\"]")
		funcMap["current_user"] = func() *users.User {
			return utils.GetCurrentUser(req)
		}

		//funcMap["current_post"] = func() *users.User {
		//	return utils.GetCurrentPost(req)
		//}

		return funcMap
	}

	return view
}
