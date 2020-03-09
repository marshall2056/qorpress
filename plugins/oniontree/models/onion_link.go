package models

import (
	"strings"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"

	"github.com/qorpress/qorpress/core/validations"
)

type OnionLink struct {
	gorm.Model
	URL            string
	Name           string `gorm:"size:255;unique" json:"href" yaml:"href"`
	Healthy        bool   `json:"healthy" yaml:"healthy"`
	OnionServiceID uint   `json:"-" yaml:"-"`
}

func (ol OnionLink) Validate(db *gorm.DB) {
	if strings.TrimSpace(ol.URL) == "" {
		db.AddError(validations.NewError(ol, "URL", "URL can not be empty"))
	}
}

func (ol *OnionLink) BeforeCreate() (err error) {
	log.Printf("======> New onion link: %#v\n", ol.URL)
	return nil
}
