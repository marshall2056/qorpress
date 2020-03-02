package posts

import (
	"net/http"
	"log"
	"strings"
	"strconv"

	// "github.com/k0kubun/pp"
	"github.com/qorpress/render"
	// pageable "github.com/BillSJC/gorm-pageable"
	"github.com/biezhi/gorm-paginator/pagination"

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
		codes          = strings.Split(utils.URLParam("code", req), "_")
		postCode    = codes[0]
		tx             = utils.GetDB(req)
	)

	if tx.Preload("Category").Where(&posts.Post{Code: postCode}).First(&post).RecordNotFound() {
		http.Redirect(w, req, "/", http.StatusFound)
	}

	// tx.Where(&posts.Post{ID: post.ID}).First(&post)
	tx.First(&post)
	ctrl.View.Execute("show", map[string]interface{}{"CurrentVariation": post}, req, w)
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
        pages = []string{"20"}
    }

    page, _ := strconv.Atoi(pages[0])
    limit, _ := strconv.Atoi(limits[0])

	db := tx.Where(&posts.Post{CategoryID: category.ID})

	/*
	resultSet := make([]*posts.Post,0)

    handler := tx.Where(&posts.Post{CategoryID: category.ID})
    //use PageQuery to get data
    resp,err := pageable.PageQuery(page, limit, handler, &resultSet)

	//Here are the response
	pp.Println(resp.PageNow)    //PageNow: current page of query
	pp.Println(resp.PageCount)  //PageCount: total page of the query
	pp.Println(resp.RawCount)   //RawCount: total raw of query
	pp.Println(resp.RawPerPage) //RawPerPage: rpp
	pp.Println(resp.ResultSet)  //ResultSet: result data
    pp.Println(resultSet)          //the same as resp.ResultSet and have the raw type
	pp.Println(resp.FirstPage)  //FirstPage: if the result is the first page
	pp.Println(resp.LastPage)   //LastPage: if the result is the last page
	pp.Println(resp.Empty)

    // you can use both resp.ResultSet or the resultSet you input to access the result
    // handle error
    if err != nil {
        panic(err)
    }
	*/
	p := pagination.Paging(&pagination.Param{
	    DB:      db,
	    Page:    page,
	    Limit:   limit,
	    // OrderBy: []string{"id desc"},
	}, &Posts)

	lastPage := (p.Page >= p.TotalPage)

	// tx.Where(&posts.Post{CategoryID: category.ID}).Find(&Posts)

	ctrl.View.Execute("category", map[string]interface{}{
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
