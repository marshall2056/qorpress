package posts

import (
	"net/http"
	"strings"
	"strconv"
	"log"
	"fmt"

	"github.com/k0kubun/pp"
	"github.com/qorpress/render"
	"github.com/acoshift/paginate"
	"github.com/qorpress/gorm-paginator/pagination"

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
		tags 		[]posts.Tag
		codes          = strings.Split(utils.URLParam("code", req), "_")
		postCode    = codes[0]
		tx             = utils.GetDB(req)
	)

	if tx.Preload("Category").Where(&posts.Post{Code: postCode}).First(&post).RecordNotFound() {
		http.Redirect(w, req, "/", http.StatusFound)
	}

	tx.First(&post)
	// tx.Where(&posts.Post{ID: post.ID}).First(&post)

	// get tags for post
	query := fmt.Sprintf(`SELECT T.Name FROM (POSTS S, TAGS T)
	INNER JOIN POST_TAGS ST ON S.ID = ST.POST_ID 
	INNER JOIN TAGS ON ST.TAG_ID = T.ID
	WHERE S.ID=%d
	GROUP BY T.NAME`, post.ID)
	tx.Raw(query).Scan(&tags)	

	ctrl.View.Execute("show", map[string]interface{}{
		"CurrentVariation": post,
		"Tags": tags,
	}, req, w)
}

func (ctrl Controller) Tag(w http.ResponseWriter, req *http.Request) {
	var (
		tag posts.Tag
		// Posts []posts.Post
		tx       = utils.GetDB(req)
	)
	if tx.Where("name = ?", utils.URLParam("name", req)).First(&tag).RecordNotFound() {
		http.Redirect(w, req, "/", http.StatusFound)
	}
	tx.First(&tag)
	ctrl.View.Execute("tag", map[string]interface{}{
		"Tag": tag,
	}, req, w)	
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

	var ok bool
	var pages, limits []string
	pages, ok = req.URL.Query()["page"]
  	if !ok || len(pages[0]) < 1 {
        log.Println("Url Param 'page' is missing")
        pages = []string{"0"}
    }

	limits, ok = req.URL.Query()["limit"]
  	if !ok || len(pages[0]) < 1 {
        log.Println("Url Param 'limit' is missing")
        limits = []string{"20"}
    }

    page, _ := strconv.Atoi(pages[0])
    limit, _ := strconv.Atoi(limits[0])

	db := tx.Where(&posts.Post{CategoryID: category.ID})

	p := pagination.Paging(&pagination.Param{
	    DB:      db,
	    Page:    page,
	    Limit:   limit,
	    // OrderBy: []string{"id desc"},
	}, &Posts)

	lastPage := (p.Page >= p.TotalPage)

	pn := paginate.New(int64(page), int64(p.Limit), int64(p.TotalRecord))

	pp.Println(pn)

	ctrl.View.Execute("category", map[string]interface{}{
		"Pagination": pn,
		"CategoryID": category.ID, 
		"CategoryName": category.Name, 
		"Posts": Posts, 
		"TotalRecord": p.TotalRecord,
		"TotalPage": p.TotalPage,
		"Offset": p.Offset,
		"Limit": p.Limit,
		"Page": p.Page,
		"PrevPage": p.PrevPage,
		"NextPage": p.NextPage,	
		"LastPage": lastPage,	
	}, req, w)
}
