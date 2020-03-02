package posts

import (
	"log"
	"strings"
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/qorpress/l10n"
	"github.com/qorpress/sorting"
	qor_seo "github.com/qorpress/seo"
	qor_slug "github.com/qorpress/slug"
	"github.com/qorpress/validations"
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

func (t Tag) Validate(db *gorm.DB) {
	if strings.TrimSpace(t.Name) == "" {
		db.AddError(validations.NewError(t, "Name", "Name can not be empty"))
	}
}

func (t Tag) DefaultPath() string {
	if len(t.Name) > 0 {
		return fmt.Sprintf("/tag/%s", t.Name)
	}
	return "/"
}


func (t *Tag) BeforeCreate() (err error) {
	// t.LanguageCode = "en-US"
	log.Printf("======> New tag: %#v\n", t.Name)
	// to do: check that the # is the prefix of the string
	return nil
}

//func (t Tag) GetSEO() *qor_seo.SEO {
//	return SEOCollection.GetSEO("Tag")
//}