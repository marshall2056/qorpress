package posts

import (
	"strings"

	"github.com/jinzhu/gorm"

	"github.com/qorpress/qorpress/internal/validations"
)

//go:generate gp-extender -structs Tag -output tag-funcs.go
type Comment struct {
	gorm.Model
	Content string `gorm:"type:mediumtext" json:"content"`
}

func (c Comment) Validate(db *gorm.DB) {
	if strings.TrimSpace(c.Content) == "" {
		db.AddError(validations.NewError(c, "Name", "Comment can not be empty"))
	}
}

func (c *Comment) BeforeCreate() (err error) {
	// check if spam (black list, maybe a detector)
	// log.Printf("======> New comment: %#v\n", c.Content)
	return nil
}
