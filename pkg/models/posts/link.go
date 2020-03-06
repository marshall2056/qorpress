package posts

import (
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/qorpress/validations"
	log "github.com/sirupsen/logrus"
)

//go:generate gp-extender -structs Link -output link-funcs.go
type Link struct {
	gorm.Model
	URL      string `gorm:"index:name"`
	Name     string
	Title    string `gorm:"type:mediumtext"`
	ImageUrl string
}

func (l Link) Validate(db *gorm.DB) {
	if strings.TrimSpace(l.URL) == "" {
		db.AddError(validations.NewError(l, "URL", "URL can not be empty"))
	}
}

func (l *Link) BeforeCreate() (err error) {
	log.Printf("======> New link: %#v\n", l.URL)
	return nil
}
