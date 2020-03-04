package posts

import (
	"log"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/qorpress/validations"
)

//go:generate gp-extender -structs Tag -output tag-funcs.go
type Comment struct {
	ID      uint   `gorm:"primary_key" json:"id"`
	Content string `json:"name"`
}

func (t Tag) Validate(db *gorm.DB) {
	if strings.TrimSpace(t.Content) == "" {
		db.AddError(validations.NewError(t, "Name", "Comment can not be empty"))
	}
}

func (t *Tag) BeforeCreate() (err error) {
	log.Printf("======> New comment: %#v\n", t.Content)
	return nil
}
