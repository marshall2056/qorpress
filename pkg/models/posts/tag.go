package posts

import (
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/qorpress/qorpress/internal/l10n"
	qor_seo "github.com/qorpress/qorpress/internal/seo"
	qor_slug "github.com/qorpress/qorpress/internal/slug"
	"github.com/qorpress/qorpress/internal/sorting"
	"github.com/qorpress/qorpress/internal/validations"
	log "github.com/sirupsen/logrus"
)

//go:generate gp-extender -structs Tag -output tag-funcs.go
type Tag struct {
	ID           uint            `gorm:"primary_key" json:"id"`
	Name         string          `json:"name" gorm:"index:name"`
	Hashtag      string          `json:"hashtag"`
	NameWithSlug qor_slug.Slug   `l10n:"sync" json:"name_with_slug" gorm:"index:name_with_slug"`
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

func (t *Tag) SetLanguageCode(code string) {
	t.LanguageCode = code
}

func (t *Tag) BeforeCreate() (err error) {
	log.Printf("======> New tag: %#v\n", t.Name)
	// to do: check that the # is the prefix of the string
	return nil
}

//func (t Tag) GetSEO() *qor_seo.SEO {
//	return SEOCollection.GetSEO("Tag")
//}
