package models

import (
	"github.com/jinzhu/gorm"
)

//go:generate gp-extender -structs Comment -output comment-funcs.go
type Comment struct {
	gorm.Model
	Content string `gorm:"type:mediumtext"`
	PostID  uint
}
