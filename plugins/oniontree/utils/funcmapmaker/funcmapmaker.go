package funcmapmaker

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/qorpress/qorpress-contrib/oniontree/models"
	"github.com/qorpress/qorpress/core/render"
	"github.com/qorpress/qorpress/pkg/models/posts"
	"github.com/qorpress/qorpress/pkg/utils"
)

func AddFuncMapMaker(view *render.Render) *render.Render {
	oldFuncMapMaker := view.FuncMapMaker
	view.FuncMapMaker = func(render *render.Render, req *http.Request, w http.ResponseWriter) template.FuncMap {
		funcMap := template.FuncMap{}
		if oldFuncMapMaker != nil {
			funcMap = oldFuncMapMaker(render, req, w)
		}

		funcMap["get_oniontree_categories"] = func() (categories []models.Category) {
			utils.GetDB(req).Find(&categories)
			return
		}

		funcMap["get_opniontree_tags"] = func(postId uint) (tags []posts.Tag) {
			query := fmt.Sprintf(`SELECT T.* FROM (POST_TAGS ST, TAGS T) WHERE ST.POST_ID=%d AND ST.tag_id=t.id`, postId)
			utils.GetDB(req).Raw(query).Scan(&tags)
			fmt.Println(query)
			return
		}

		funcMap["get_opniontree_category_tags"] = func(catId uint) (tags []posts.Tag) {
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

		return funcMap
	}

	return view
}
