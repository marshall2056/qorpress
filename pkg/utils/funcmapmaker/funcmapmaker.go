package funcmapmaker

import (
	"html/template"
	"net/http"

	"github.com/qorpress/action_bar"
	"github.com/qorpress/i18n/inline_edit"
	"github.com/qorpress/qor"
	"github.com/qorpress/render"
	"github.com/qorpress/session"
	"github.com/qorpress/session/manager"
	"github.com/qorpress/widget"

	"github.com/qorpress/qorpress-example/pkg/app/admin"
	"github.com/qorpress/qorpress-example/pkg/config/i18n"
	"github.com/qorpress/qorpress-example/pkg/models/posts"
	"github.com/qorpress/qorpress-example/pkg/models/seo"
	"github.com/qorpress/qorpress-example/pkg/models/users"
	"github.com/qorpress/qorpress-example/pkg/utils"
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
			return admin.ActionBar.Actions(action_bar.Action{Name: "Edit SEO", Link: seo.SEOCollection.SEOSettingURL("/help")}).Render(w, req)
		}

		funcMap["render_seo_tag"] = func() template.HTML {
			return seo.SEOCollection.Render(&qor.Context{DB: utils.GetDB(req)}, "Default Page")
		}

		funcMap["get_categories"] = func() (categories []posts.Category) {
			utils.GetDB(req).Find(&categories)
			return
		}

		funcMap["current_locale"] = func() string {
			return utils.GetCurrentLocale(req)
		}

		funcMap["current_user"] = func() *users.User {
			return utils.GetCurrentUser(req)
		}

        funcMap["Iterate"] = func(count uint) []uint {
            var i uint
            var Items []uint
            for i = 0; i < (count); i++ {
                Items = append(Items, i)
            }
            return Items
        }

		return funcMap
	}

	return view
}
