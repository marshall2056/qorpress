package models

import (
	"log"

	"github.com/qorpress/l10n"
	"github.com/qorpress/sorting"
	qor_seo "github.com/qorpress/seo"
	qor_slug "github.com/qorpress/slug"
)

//go:generate gp-extender -structs Tag -output tag-funcs.go
type Tag struct {
	ID      uint `gorm:"primary_key" json:"id"`
	Name    string `json:"name"`
	Hashtag string `json:"hashtag"`
	NameWithSlug qor_slug.Slug `l10n:"sync" json:"name_with_slug"`
	Seo          qor_seo.Setting `json:"seo"`
	l10n.Locale
	sorting.Sorting
}

func (t *Tag) BeforeCreate() (err error) {
	t.LanguageCode = "en-US"
	log.Printf("======> New tag: %#v\n", t.Name)
	// to do: check that the # is the prefix of the string
	return nil
}

func (t Tag) GetSEO() *qor_seo.SEO {
	return SEOCollection.GetSEO("Tag")
}