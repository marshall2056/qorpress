package models

import (
	"github.com/qorpress/page_builder"
	"github.com/qorpress/publish2"
	qor_seo "github.com/qorpress/seo"
	qor_slug "github.com/qorpress/slug"
)

type Page struct {
	page_builder.Page
	NameWithSlug qor_slug.Slug `l10n:"sync"`
	Seo          qor_seo.Setting
	publish2.Version
	publish2.Schedule
	publish2.Visible
}

func (p Page) GetSEO() *qor_seo.SEO {
	return SEOCollection.GetSEO("Page")
}

/*
//go:generate gp-extender -structs Page -output page-funcs.go
type Page struct {
	ID         uint       `gorm:"primary_key"`
	Categories []Category `gorm:"many2many:category_page"`
	Tags       []Tag      `gorm:"many2many:tag_page"`
	Title      string
	Slug       string `gorm:"unique"`
	Body       string `gorm:"type:longtext"`
	Summary    string `gorm:"type:longtext"`
	Images     []Image
	Documents  []Document
	Videos     []Video
	Links      []Link
	Type       string
	Created    int32
	Updated    int32
}
*/
