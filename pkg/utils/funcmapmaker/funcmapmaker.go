package funcmapmaker

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/gopress/internal/action_bar"
	"github.com/gopress/internal/i18n/inline_edit"
	"github.com/gopress/internal/qor"
	"github.com/gopress/internal/render"
	"github.com/gopress/internal/session"
	"github.com/gopress/internal/session/manager"
	"github.com/gopress/internal/widget"

	// "github.com/k0kubun/pp"
	"github.com/gopress/qorpress/pkg/app/admin"
	"github.com/gopress/qorpress/pkg/config/i18n"
	"github.com/gopress/qorpress/pkg/models/posts"
	"github.com/gopress/qorpress/pkg/models/seo"
	"github.com/gopress/qorpress/pkg/models/users"
	"github.com/gopress/qorpress/pkg/utils"
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

		funcMap["raw"] = func(str string) template.HTML {
			return template.HTML(utils.HTMLSanitizer.Sanitize(str))
		}

		funcMap["flashes"] = func() []session.Message {
			return manager.SessionManager.Flashes(w, req)
		}

		// Add `action_bar` method
		funcMap["render_action_bar"] = func() template.HTML {
			return admin.ActionBar.Actions(action_bar.Action{
				Name: "Edit SEO",
				Link: seo.SEOCollection.SEOSettingURL("/help"),
			}).Render(w, req)
		}

		funcMap["render_seo_tag"] = func() template.HTML {
			return seo.SEOCollection.Render(&qor.Context{DB: utils.GetDB(req)}, "Default Page")
		}

		funcMap["get_categories"] = func() (categories []posts.Category) {
			utils.GetDB(req).Find(&categories)
			return
		}

		funcMap["get_post_tags"] = func(postId uint) (tags []posts.Tag) {
			query := fmt.Sprintf(`SELECT T.* FROM (POST_TAGS ST, TAGS T) WHERE ST.POST_ID=%d AND ST.tag_id=t.id`, postId)
			utils.GetDB(req).Raw(query).Scan(&tags)
			// pp.Println("tags:", tags)
			fmt.Println(query)
			return
		}

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

		funcMap["current_locale"] = func() string {
			return utils.GetCurrentLocale(req)
		}

		funcMap["current_user"] = func() *users.User {
			return utils.GetCurrentUser(req)
		}

		return funcMap
	}

	return view
}
