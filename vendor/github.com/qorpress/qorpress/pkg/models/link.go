package models

import (
	"github.com/jinzhu/gorm"
)

//go:generate gp-extender -structs Link -output link-funcs.go
type Link struct {
	gorm.Model
	Url      string
	Title    string `gorm:"type:mediumtext"`
	ImageUrl string
	PostID   uint
}
