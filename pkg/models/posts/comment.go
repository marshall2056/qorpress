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

func (c Comment) Validate(db *gorm.DB) {
	if strings.TrimSpace(c.Content) == "" {
		db.AddError(validations.NewError(c, "Name", "Comment can not be empty"))
	}
}

func (c *Comment) BeforeCreate() (err error) {
	log.Printf("======> New comment: %#v\n", c.Content)
	return nil
}
