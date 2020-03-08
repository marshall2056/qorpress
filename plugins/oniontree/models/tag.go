package models

import (
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"

	"github.com/qorpress/qorpress/core/l10n"
	qor_seo "github.com/qorpress/qorpress/core/seo"
	qor_slug "github.com/qorpress/qorpress/core/slug"
	"github.com/qorpress/qorpress/core/sorting"
	"github.com/qorpress/qorpress/core/validations"
)

type OnionTag struct {
	gorm.Model
	Name         string          `gorm:"size:32;unique" json:"name" yaml:"name"`
	NameWithSlug qor_slug.Slug   `l10n:"sync" json:"name_with_slug" gorm:"index:name_with_slug"`
	Seo          qor_seo.Setting `json:"seo"`
	l10n.Locale
	sorting.Sorting
}

func (o OnionTag) Validate(db *gorm.DB) {
	if strings.TrimSpace(o.Name) == "" {
		db.AddError(validations.NewError(o, "Name", "Name can not be empty"))
	}
}

func (o OnionTag) DefaultPath() string {
	if len(o.Name) > 0 {
		return fmt.Sprintf("/onion-tag/%s", o.Name)
	}
	return "/"
}

func (o *OnionTag) SetLanguageCode(code string) {
	o.LanguageCode = code
}

func (o *OnionTag) BeforeCreate() (err error) {
	log.Printf("======> New Onion-Tag: %#v\n", o.Name)
	// to do: check that the # is the prefix of the string
	return nil
}
