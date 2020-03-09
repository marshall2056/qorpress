package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/acoshift/paginate"
	"github.com/biezhi/gorm-paginator/pagination"
	"github.com/k0kubun/pp"

	"github.com/qorpress/qorpress-contrib/oniontree/models"
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
		Services []models.OnionService
		tx       = utils.GetDB(req)
	)
	tx.Preload("OnionTag").Find(&Services)
	ctrl.View.Execute("index", map[string]interface{}{}, req, w)
}

// Show post show page
func (ctrl Controller) Show(w http.ResponseWriter, req *http.Request) {
	var (
		service     models.OnionService
		tags        []models.OnionTag
		codes       = strings.Split(utils.URLParam("code", req), "_")
		serviceCode = codes[0]
		tx          = utils.GetDB(req)
	)

	if tx.Preload("OnionTag").Where(&models.OnionService{Code: serviceCode}).First(&service).RecordNotFound() {
		http.Redirect(w, req, "/", http.StatusFound)
	}

	tx.First(&service)
	pp.Println("service: ", service)

	ctrl.View.Execute("show", map[string]interface{}{
		"CurrentVariation": service,
		"Tags":             tags,
	}, req, w)
}

func (ctrl Controller) Tag(w http.ResponseWriter, req *http.Request) {
	var (
		tag      models.OnionTag
		Services []models.OnionService
		tx       = utils.GetDB(req)
	)

	if tx.Where("name = ?", utils.URLParam("code", req)).First(&tag).RecordNotFound() {
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

	offset := page * limit
	query := fmt.Sprintf(`SELECT P.* FROM (POSTS P, POST_TAGS PT) WHERE PT.POST_ID=P.ID AND PT.tag_id=%d LIMIT %d OFFSET %d`, tag.ID, limit, offset)
	pp.Println(query)

	tx.Raw(query).Scan(&Services)

	ctrl.View.Execute("tag", map[string]interface{}{
		"Tag":      tag,
		"Services": Services,
	}, req, w)
}

// Category category show page
func (ctrl Controller) Category(w http.ResponseWriter, req *http.Request) {
	var (
		category models.OnionCategory
		Services []models.OnionService
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

	db := tx.Where(&models.OnionService{CategoryID: category.ID})

	p := pagination.Paging(&pagination.Param{
		DB:      db,
		Page:    page,
		Limit:   limit,
		OrderBy: []string{"id desc"},
	}, &Services)

	lastPage := (p.Page >= p.TotalPage)

	pn := paginate.New(int64(page), int64(p.Limit), int64(p.TotalRecord))

	pp.Println(pn)

	ctrl.View.Execute("category", map[string]interface{}{
		"Pagination":   pn,
		"CategoryID":   category.ID,
		"CategoryName": category.Name,
		"Services":     Services,
		"TotalRecord":  p.TotalRecord,
		"TotalPage":    p.TotalPage,
		"Offset":       p.Offset,
		"Limit":        p.Limit,
		"Page":         p.Page,
		"PrevPage":     p.PrevPage,
		"NextPage":     p.NextPage,
		"LastPage":     lastPage,
	}, req, w)
}
