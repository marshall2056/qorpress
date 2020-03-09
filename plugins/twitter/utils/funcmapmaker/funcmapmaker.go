package funcmapmaker

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/qorpress/qorpress-contrib/twitter/models"
	"github.com/qorpress/qorpress/core/render"
	"github.com/qorpress/qorpress/pkg/utils"
)

// jeudi 10:00 madame 

func AddFuncMapMaker(view *render.Render) *render.Render {
	oldFuncMapMaker := view.FuncMapMaker
	view.FuncMapMaker = func(render *render.Render, req *http.Request, w http.ResponseWriter) template.FuncMap {
		funcMap := template.FuncMap{}
		if oldFuncMapMaker != nil {
			funcMap = oldFuncMapMaker(render, req, w)
		}

		funcMap["get_tweets"] = func() (tweets []models.TwitterTweet) {
			utils.GetDB(req).Find(&tweets)
			return
		}

		funcMap["get_twitter_profile"] = func() (profile models.TwitterProfile) {
			query := fmt.Sprintf(`SELECT T.* FROM (POST_TAGS ST, TAGS T) WHERE ST.tag_id=t.id`)
			utils.GetDB(req).Raw(query).Scan(&profile)
			fmt.Println(query)
			return
		}

		return funcMap
	}

	return view
}
